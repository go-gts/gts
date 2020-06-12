package gts

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/go-ascii/ascii"
	"github.com/go-pars/pars"
)

// Feature represents a single feature within a feature table.
type Feature struct {
	Key        string
	Location   Location
	Qualifiers Values

	order map[string]int
}

func listQualifiers(f Feature) []QualifierIO {
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

	qfs := make([]QualifierIO, 0, len(names))

	for _, name := range names {
		for _, value := range f.Qualifiers[name] {
			qfs = append(qfs, QualifierIO{name, value})
		}
	}

	return qfs
}

// Filter represents a filtering function for a Feature.
type Filter func(f Feature) bool

// TrueFilter always returns true.
func TrueFilter(f Feature) bool { return true }

// FalseFilter always return false.
func FalseFilter(f Feature) bool { return false }

// Key returns true if the key of a feature matches the given key string.
func Key(key string) Filter {
	if key == "" {
		return TrueFilter
	}
	return func(f Feature) bool { return f.Key == key }
}

// Qualifier returns true if the value for the given qualifier name matches the
// given regex expression.
func Qualifier(name, query string) (Filter, error) {
	re, err := regexp.Compile(query)
	if err != nil {
		return FalseFilter, err
	}
	return func(f Feature) bool {
		if values, ok := f.Qualifiers[name]; ok {
			for _, value := range values {
				if re.MatchString(value) {
					return true
				}
			}
		}
		return false
	}, nil
}

// And creates a Filter which will return true if all of the filters given to
// And return true for the Feature.
func And(filters ...Filter) Filter {
	return func(f Feature) bool {
		for _, filter := range filters {
			if !filter(f) {
				return false
			}
		}
		return true
	}
}

// Or creates a Filter which will return true if any of the filters given to
// Or return true for the Feature.
func Or(filters ...Filter) Filter {
	return func(f Feature) bool {
		for _, filter := range filters {
			if filter(f) {
				return true
			}
		}
		return false
	}
}

// Not creates a Filter which will return true if the given Filter does not
// return true for the Feature.
func Not(filter Filter) Filter {
	return func(f Feature) bool {
		return !filter(f)
	}
}

func selectorShift(s string) (string, string) {
	esc := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			esc = true
		case '/':
			if !esc {
				return s[:i], s[i+1:]
			}
		default:
			esc = false
		}
	}
	return s, ""
}

func toQualifier(s string) (Filter, error) {
	if i := strings.IndexByte(s, '='); i >= 0 {
		return Qualifier(s[:i], s[i+1:])
	}
	return Qualifier(s, "")
}

// Selector creates a Filter which will return true if the Feature matches the
// given selector string.
func Selector(sel string) (Filter, error) {
	head, tail := selectorShift(sel)
	filter := Key(head)
	for tail != "" {
		head, tail = selectorShift(tail)
		qfs, err := toQualifier(head)
		if err != nil {
			return FalseFilter, err
		}
		filter = And(filter, qfs)
	}
	return filter, nil
}

// FeatureTable represents a table of features.
type FeatureTable []Feature

// Filter returns a FeatureTable containing the features that match the given
// Filter within this FeatureTable.
func (ff FeatureTable) Filter(filter Filter) FeatureTable {
	gg := FeatureTable{}
	for _, f := range ff {
		if filter(f) {
			gg = append(gg, f)
		}
	}
	return gg
}

// Insert takes the given Feature and inserts it into the sorted position in
// the FeatureTable.
func (ff FeatureTable) Insert(f Feature) FeatureTable {
	i := 0
	for i < len(ff) && ff[i].Key == "source" {
		i++
	}
	if f.Key != "source" {
		i += sort.Search(len(ff[i:]), func(j int) bool {
			return f.Location.Less(ff[i+j].Location)
		})
	}

	ff = append(ff, Feature{})
	copy(ff[i+1:], ff[i:])
	ff[i] = f

	return ff
}

// Format creates a FeatureTableFormatter object for the qualifier with the
// given prefix and depth. If the Feature object was created by parsing some
// input, the qualifier values will be in the same order as in the input
// source. The exception to this rule is the `translation` qualifier which will
// always be written last. Qualifiers given during runtime will be sorted in
// ascending alphabetical order and written after the qualifiers present in the
// source.
func (ff FeatureTable) Format(prefix string, depth int) FeatureTableFormatter {
	return FeatureTableFormatter{ff, prefix, depth}
}

// FeatureTableFormatter formats a Feature object with the given prefix and depth.
type FeatureTableFormatter struct {
	Table  FeatureTable
	Prefix string
	Depth  int
}

// String satisfies the fmt.Stringer interface.
func (ftf FeatureTableFormatter) String() string {
	builder := strings.Builder{}
	for i, f := range ftf.Table {
		if i != 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(ftf.Prefix)
		builder.WriteString(f.Key)
		length := len(ftf.Prefix) + len(f.Key)

		padding := strings.Repeat(" ", ftf.Depth-length)
		prefix := ftf.Prefix + strings.Repeat(" ", ftf.Depth-len(ftf.Prefix))

		builder.WriteString(padding)
		builder.WriteString(f.Location.String())

		for _, q := range listQualifiers(f) {
			builder.WriteByte('\n')
			builder.WriteString(q.Format(prefix).String())
		}
	}
	return builder.String()
}

// WriteTo satisfies the io.WriteTo interface.
func (ftf FeatureTableFormatter) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, ftf.String())
	return int64(n), err
}

type keyline struct {
	pre int
	key string
	pst int
	loc Location
}

func featureKeylineParser(prefix string, depth int) pars.Parser {
	word := pars.Word(ascii.IsSnake)
	p := []byte(prefix)
	return func(state *pars.State, result *pars.Result) error {
		if err := state.Request(len(p)); err != nil {
			return err
		}
		if !bytes.Equal(state.Buffer(), p) {
			return pars.NewError(fmt.Sprintf("expected %q", prefix), state.Position())
		}
		state.Advance()
		if err := word(state, result); err != nil {
			return err
		}
		key := string(result.Token)
		for i := 0; i < depth-len(prefix+key); i++ {
			c, err := pars.Next(state)
			if err != nil {
				return err
			}
			if c != ' ' {
				return pars.NewError("wanted indent", state.Position())
			}
			state.Advance()
		}
		if err := LocationParser(state, result); err != nil {
			return err
		}
		loc := result.Value.(Location)
		if err := pars.EOL(state, result); err != nil {
			return err
		}
		result.SetValue(keyline{0, key, 0, loc})
		return nil
	}
}

// FeatureTableParser attempts to match an INSDC feature table.
func FeatureTableParser(prefix string) pars.Parser {
	firstParser := pars.Seq(
		prefix, pars.Spaces,
		pars.Word(ascii.IsSnake), pars.Spaces,
		LocationParser, pars.EOL,
	).Map(func(result *pars.Result) error {
		children := result.Children
		pre := len(children[1].Token)
		key := string(children[2].Token)
		pst := len(children[3].Token)
		loc := children[4].Value.(Location)
		result.SetValue(keyline{pre, key, pst, loc})
		return nil
	})

	return func(state *pars.State, result *pars.Result) error {
		if err := firstParser(state, result); err != nil {
			return err
		}
		tmp := result.Value.(keyline)
		pre, key, pst, loc := tmp.pre, tmp.key, tmp.pst, tmp.loc
		depth := pre + len(key) + pst

		keylineParser := featureKeylineParser(prefix+strings.Repeat(" ", pre), depth)

		qualifierParser := QualifierParser(prefix + strings.Repeat(" ", depth))
		qualifiersParser := pars.Many(pars.Seq(qualifierParser, pars.EOL).Child(0))

		// Does not return error by definition.
		qualifiersParser(state, result)

		qfs := Values{}
		order := make(map[string]int)

		for _, child := range result.Children {
			name, value := child.Value.(QualifierIO).Unpack()
			qfs.Add(name, value)
			if _, ok := order[name]; name != "translation" && !ok {
				order[name] = len(order)
			}
		}

		ff := []Feature{{
			Key:        key,
			Location:   loc,
			Qualifiers: qfs,
			order:      order,
		}}

		for keylineParser(state, result) == nil {
			tmp := result.Value.(keyline)
			key, loc := tmp.key, tmp.loc

			// Does not return error by definition.
			qualifiersParser(state, result)

			qfs := Values{}
			order := make(map[string]int)

			for _, child := range result.Children {
				name, value := child.Value.(QualifierIO).Unpack()
				qfs.Add(name, value)
				if _, ok := order[name]; name != "translation" && !ok {
					order[name] = len(order)
				}
			}

			ff = append(ff, Feature{
				Key:        key,
				Location:   loc,
				Qualifiers: qfs,
				order:      order,
			})
		}

		result.SetValue(FeatureTable(ff))
		return nil
	}
}
