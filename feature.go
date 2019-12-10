package gts

import (
	"io"
	"sort"
	"strings"

	ascii "gopkg.in/ktnyt/ascii.v1"
	pars "gopkg.in/ktnyt/pars.v2"
)

// Feature represents a single feaute within a feature table.
type Feature struct {
	Key        string
	Location   Location
	Qualifiers Values
	order      map[string]int
}

// NewFeature creates a new feature.
func NewFeature(key string, loc Location, qfs Values) Feature {
	return Feature{key, loc, qfs, nil}
}

// Translation will return the translation of the feature if available. it will
// return nil otherwise.
func (f Feature) Translation() Sequence {
	if values := f.Qualifiers.Get("translation"); len(values) != 0 {
		s := values[0]
		return Seq(strings.ReplaceAll(s, "\n", ""))
	}
	return nil
}

// Format creates a FeatureFormatter object for the qualifier with the given
// prefix and depth. If the Feature object was created by parsing some input,
// the qualifier values will be in the same order as in the input source. The
// exception to this rule is the `translation` qualifier which will always be
// written last. Qualifiers given during runtime will be sorted in ascending
// alphabetical order and written after the qualifiers present in the source.
func (f Feature) Format(prefix string, depth int) FeatureFormatter {
	return FeatureFormatter{f, prefix, depth}
}

// FeatureFormatter will format a Feature object with the given prefix and
// depth.
type FeatureFormatter struct {
	Feature Feature
	Prefix  string
	Depth   int
}

// String satisfies the fmt.Stringer interface.
func (ff FeatureFormatter) String() string {
	builder := strings.Builder{}
	builder.WriteString(ff.Prefix)
	builder.WriteString(ff.Feature.Key)

	padding := strings.Repeat(" ", ff.Depth-builder.Len())
	prefix := ff.Prefix + strings.Repeat(" ", ff.Depth-len(ff.Prefix))

	builder.WriteString(padding)
	builder.WriteString(ff.Feature.Location.String())

	ordered := make([]string, len(ff.Feature.order))
	remains := []string{}

	hasTranslate := false

	for name := range ff.Feature.Qualifiers {
		index, ok := ff.Feature.order[name]
		switch {
		case ok:
			ordered[index] = name
		case name == "translation":
			hasTranslate = true
		default:
			remains = append(remains, name)
		}
	}

	for i, name := range ordered {
		if name == "" {
			ordered = append(ordered[:i], ordered[i+1:]...)
		}
	}

	sort.Strings(remains)

	names := append(ordered, remains...)

	if hasTranslate {
		names = append(names, "translation")
	}

	for _, name := range names {
		for _, value := range ff.Feature.Qualifiers[name] {
			q := Qualifier{name, value}
			builder.WriteByte('\n')
			builder.WriteString(q.Format(prefix).String())
		}
	}

	return builder.String()
}

// WriteTo satisfies the io.WriteTo interface.
func (ff FeatureFormatter) WriteTo(w io.Writer) (int, error) {
	return w.Write([]byte(ff.String()))
}

type keyline struct {
	key string
	pad int
	loc Location
}

// FeatureParser will attempt to match a single feature.
func FeatureParser(prefix string) pars.Parser {
	keylineParser := pars.Seq(
		prefix, pars.Spaces,
		pars.Word(ascii.IsSnake), pars.Spaces,
		LocationParser, pars.EOL,
	).Map(func(result *pars.Result) error {
		children := result.Children
		pad := 0
		pad += len(children[1].Token)
		key := string(children[2].Token)
		pad += len(key)
		pad += len(children[3].Token)
		loc := children[4].Value.(Location)
		result.SetValue(keyline{key, pad, loc})
		return nil
	})

	return func(state *pars.State, result *pars.Result) error {
		state.Request(1)
		if err := keylineParser(state, result); err != nil {
			return err
		}
		tmp := result.Value.(keyline)
		key := tmp.key
		pad := tmp.pad
		loc := tmp.loc

		qualifierParser := QualifierParser(prefix + strings.Repeat(" ", pad))
		qualifiersParser := pars.Many(pars.Seq(qualifierParser, pars.EOL).Child(0))

		// Does not return error by definition.
		qualifiersParser(state, result)

		qualifiers := Values{}
		order := make(map[string]int)

		for _, child := range result.Children {
			q := child.Value.(Qualifier)
			qualifiers.Add(q.Name, q.Value)
			if _, ok := order[q.Name]; q.Name != "translation" && !ok {
				order[q.Name] = len(order)
			}
		}

		result.SetValue(Feature{key, loc, qualifiers, order})
		return nil
	}
}
