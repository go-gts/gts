package gts

import (
	"sort"
	"strings"

	"gopkg.in/ktnyt/ascii.v1"
	"gopkg.in/ktnyt/pars.v2"
)

// Feature represents a single feaute within a feature table.
type Feature struct {
	Key        string
	Loc        Location
	Qualifiers map[string][]string
	order      map[string]int
}

// Format will format the feature. If the feature was parsed from some input,
// the qualifier values will be in the same order as in the input source. The
// exception to this rule is the `translation` qualifier which will always be
// written last. Qualifiers given during runtime will be sorted in ascending
// alphabetical order and written after the qualifiers present in the source.
func (f Feature) Format(prefix string, depth int) string {
	builder := strings.Builder{}
	builder.WriteString(prefix)
	builder.WriteString(f.Key)

	padding := strings.Repeat(" ", depth-builder.Len())
	prefix = prefix + strings.Repeat(" ", depth-len(prefix))

	builder.WriteString(padding)
	builder.WriteString(f.Loc.String())

	ordered := make([]string, len(f.order))
	remains := []string{}

	hasTranslate := false

	for name := range f.Qualifiers {
		index, ok := f.order[name]
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
		for _, value := range f.Qualifiers[name] {
			q := Qualifier{name, value}
			builder.WriteByte('\n')
			builder.WriteString(q.Format(prefix))
		}
	}

	return builder.String()
}

func featureKeyLineParser(prefix string) pars.Parser {
	spaces := pars.Word(ascii.Is(' '))
	word := pars.Word(ascii.IsSnake)
	return pars.Seq(spaces, word, spaces, LocationParser, pars.EOL)
}

// FeatureParser will attempt to match a single feature.
func FeatureParser(prefix string) pars.Parser {
	keylineParser := featureKeyLineParser(prefix)

	return func(state *pars.State, result *pars.Result) error {
		state.Request(1)
		if err := keylineParser(state, result); err != nil {
			return err
		}
		padding := len(result.Children[0].Token) + len(result.Children[2].Token)
		key := string(result.Children[1].Token)
		loc := result.Children[3].Value.(Location)
		indent := strings.Repeat(" ", padding+len(key))

		qualifierParser := pars.Many(
			pars.Seq(
				QualifierParser(prefix+indent),
				pars.EOL,
			).Child(0),
		)

		if err := qualifierParser(state, result); err != nil {
			return err
		}

		qualifiers := Qualifiers{}
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
