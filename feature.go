package gts

import (
	"sort"
	"strings"

	ascii "gopkg.in/ktnyt/ascii.v1"
	pars "gopkg.in/ktnyt/pars.v2"
)

// Feature represents a single feaute within a feature table.
type Feature struct {
	Key        string
	Loc        Location
	Qualifiers Qualifiers
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
