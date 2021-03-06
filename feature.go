package gts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/go-ascii/ascii"
	"github.com/go-pars/pars"
)

// Feature represents a single feature within an INSDC feature table. Each
// feature has a feature key, a location, and qualifiers in the form of key
// value pairs. A single qualifier name may have multiple values. The Feature
// object can additionally store the order in which the qualifiers should
// appear when it is formatted. Regardless of the specified order, the
// `translation` qualifier will always appear last. Qualifiers whose names
// appear in the ordering map will be prioritized over those that do not. All
// qualifier names that do not appear in the ordering map will simply be
// arranged in alphabetical order.
type Feature struct {
	Key        string
	Location   Location
	Qualifiers Values
	Order      map[string]int
}

func listQualifiers(f Feature) []QualifierIO {
	ordered := make([]string, len(f.Order))
	remains := []string{}

	hasTranslate := false

	for name := range f.Qualifiers {
		index, ok := f.Order[name]
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

// Repair attempts to reconstruct features by joining features with identical
// feature keys and qualifiers which have adjacent locations.
func Repair(ff []Feature) []Feature {
	gg := make([]Feature, len(ff))
	copy(gg, ff)

	// Identify the features with similar keys and qualifiers.
	index := make(map[string][]int)
	for i, f := range gg {
		key := fmt.Sprintf("%s:%v", f.Key, f.Qualifiers)
		index[key] = append(index[key], i)
	}

	remove := []int{}
	for _, ii := range index {
		if len(ii) > 1 {
			locs := make([]Location, len(ii))
			for j, i := range ii {
				locs[j] = gg[i].Location
			}
			sort.Sort(Locations(locs))

			force := ff[ii[0]].Key == "source"
			list := LocationList{}
			for _, loc := range locs {
				list.Push(loc, force)
			}

			// DISCUSS: Should we join these locations?
			locs = list.Slice()

			// Some locations were merged.
			if len(locs) < len(ii) {
				for j := range ii {
					i := ii[j]
					if j < len(locs) {
						gg[i].Location = locs[j]
					} else {
						// Remove the excess features.
						remove = append(remove, i)
					}
				}
			}
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(remove)))

	for _, i := range remove {
		copy(gg[i:], gg[i+1:])
		gg[len(gg)-1] = Feature{}
		gg = gg[:len(gg)-1]
	}

	return gg
}

// Filter represents a filtering function for a Feature. It should return a
// boolean value upon receiveing a Feature object.
type Filter func(f Feature) bool

// And generates a new Filter which will only return true if all of the given
// filters return true for a given Feature object.
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

// Or generates a new Filter which will return true if any one of the given
// filters return true for a given Feature object.
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

// Not generates a new Filter which will return true if the given Filter
// returns false for a given Feature object.
func Not(filter Filter) Filter {
	return func(f Feature) bool {
		return !filter(f)
	}
}

// TrueFilter always returns true.
func TrueFilter(f Feature) bool { return true }

// FalseFilter always return false.
func FalseFilter(f Feature) bool { return false }

// Within returns true if the location of the feature is within the given
// bounds.
func Within(lower, upper int) Filter {
	return func(f Feature) bool {
		return LocationWithin(f.Location, lower, upper)
	}
}

// Overlap returns true if the location of the feature overlaps with the given
// bounds.
func Overlap(lower, upper int) Filter {
	return func(f Feature) bool {
		return LocationOverlap(f.Location, lower, upper)
	}
}

// Key returns true if the key of a feature matches the given key string. If
// an empty string was given, the filter will always return true.
func Key(key string) Filter {
	if key == "" {
		return TrueFilter
	}
	return func(f Feature) bool { return f.Key == key }
}

// Qualifier tests if any of the values associated with the given qualifier
// name matches the given regular expression query. If the qualifier name is
// empty, the values for every qualifier name will be tested.
func Qualifier(name, query string) (Filter, error) {
	re, err := regexp.Compile(query)
	if err != nil {
		return FalseFilter, err
	}

	if name == "" {
		return func(f Feature) bool {
			for _, vv := range f.Qualifiers {
				for _, v := range vv {
					if re.MatchString(v) {
						return true
					}
				}
			}
			return false
		}, nil
	}

	if query == "" {
		return func(f Feature) bool {
			_, ok := f.Qualifiers[name]
			return ok
		}, nil
	}

	return func(f Feature) bool {
		if vv, ok := f.Qualifiers[name]; ok {
			for _, v := range vv {
				if re.MatchString(v) {
					return true
				}
			}
		}
		return false
	}, nil
}

func shiftSelector(s string) (string, string) {
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
	switch i := strings.IndexByte(s, '='); i {
	case -1:
		return Qualifier(s, "")
	default:
		return Qualifier(s[:i], s[i+1:])
	}
}

// Selector generates a new Filter which will return true if a given Feature
// satisfies the criteria specified by the selection string. A selector in GTS
// is defined as follows:
//   [feature_key]/qualifier_name=regexp[/qualifier_name=regexp]...
// If the qualifier name is omitted, the values for every qualifier name will
// be tested.
func Selector(sel string) (Filter, error) {
	head, tail := shiftSelector(sel)
	filter := Key(head)
	for tail != "" {
		head, tail = shiftSelector(tail)
		qfs, err := toQualifier(head)
		if err != nil {
			return FalseFilter, err
		}
		filter = And(filter, qfs)
	}
	return filter, nil
}

// FeatureTable represents an INSDC feature table. Unless explicitly set, the
// order of features appearing in the FeatureTable should be in ascending order
// based on the location of the feature with the exception being sources.
type FeatureTable []Feature

// Filter returns a FeatureTable containing the features that match the given
// Filter within this FeatureTable.
func (ff FeatureTable) Filter(filter Filter) FeatureTable {
	indices := make([]int, 0, len(ff))
	for i, f := range ff {
		if filter(f) {
			indices = append(indices, i)
		}
	}
	gg := make(FeatureTable, len(indices))
	for i, index := range indices {
		gg[i] = ff[index]
	}
	return gg
}

// Len is the number of elements in the collection.
func (ff FeatureTable) Len() int {
	return len(ff)
}

// Less reports whether the element with index i should sort before the element
// with index j.
func (ff FeatureTable) Less(i, j int) bool {
	f, g := ff[i], ff[j]
	if f.Key == "source" && g.Key != "source" {
		return true
	}
	if f.Key != "source" && g.Key == "source" {
		return false
	}
	return LocationLess(f.Location, g.Location)
}

// Swap the elements with indexes i and j.
func (ff FeatureTable) Swap(i, j int) {
	ff[i], ff[j] = ff[j], ff[i]
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
			return LocationLess(f.Location, ff[i+j].Location)
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
	b := strings.Builder{}
	for i, f := range ftf.Table {
		if i != 0 {
			b.WriteByte('\n')
		}
		b.WriteString(ftf.Prefix)
		b.WriteString(f.Key)
		length := len(ftf.Prefix) + len(f.Key)

		padding := strings.Repeat(" ", ftf.Depth-length)
		prefix := ftf.Prefix + strings.Repeat(" ", ftf.Depth-len(ftf.Prefix))

		b.WriteString(padding)
		b.WriteString(f.Location.String())

		for _, q := range listQualifiers(f) {
			b.WriteByte('\n')
			b.WriteString(q.Format(prefix).String())
		}
	}
	return b.String()
}

// WriteTo satisfies the io.WriteTo interface.
func (ftf FeatureTableFormatter) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, ftf.String())
	return int64(n), err
}

var errFeatureKey = errors.New("expected feature key")

type keyline struct {
	pre int
	key string
	pst int
	loc Location
}

func featureKeylineParser(prefix string, depth int) pars.Parser {
	word := pars.Word(ascii.IsSnake).Error(errFeatureKey)
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
		if err := parseLocation(state, result); err != nil {
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
		pars.Word(ascii.IsSnake).Error(errFeatureKey), pars.Spaces,
		parseLocation, pars.EOL,
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
		qualifiersParser := pars.Many(qualifierParser)

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
			Order:      order,
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
				Order:      order,
			})
		}

		result.SetValue(FeatureTable(ff))
		return nil
	}
}
