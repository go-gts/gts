package gts

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Feature interface {
	Key() string
	Location() Location
	Values() Props
}

type BasicFeature struct {
	key  string
	loc  Location
	vals Props
}

func (f BasicFeature) Key() string {
	return f.key
}

func (f BasicFeature) Location() Location {
	return f.loc
}

func (f BasicFeature) Values() Props {
	return f.vals
}

func NewFeature(key string, loc Location, values Props) BasicFeature {
	return BasicFeature{key, loc, values}
}

type hasWithKey interface {
	WithKey(key string) Feature
}

type hasWithLocation interface {
	WithLocation(loc Location) Feature
}

type hasWithValues interface {
	WithValues(values Props) Feature
}

func WithKey(f Feature, key string) Feature {
	switch v := f.(type) {
	case hasWithKey:
		return v.WithKey(key)
	default:
		return NewFeature(key, f.Location(), f.Values())
	}
}

func WithLocation(f Feature, loc Location) Feature {
	switch v := f.(type) {
	case hasWithLocation:
		return v.WithLocation(loc)
	default:
		return NewFeature(f.Key(), loc, f.Values())
	}
}

func WithValues(f Feature, values Props) Feature {
	switch v := f.(type) {
	case hasWithValues:
		return v.WithValues(values)
	default:
		return NewFeature(f.Key(), f.Location(), values)
	}
}

// Repair attempts to reconstruct features by joining features with identical
// feature keys and values which have adjacent locations.
func Repair(ff []Feature) []Feature {
	gg := make([]Feature, len(ff))
	copy(gg, ff)

	// Identify the features with similar keys and values.
	index := make(map[string][]int)
	for i, f := range gg {
		key := fmt.Sprintf("%s:%v", f.Key(), f.Values())
		index[key] = append(index[key], i)
	}

	remove := []int{}
	for _, ii := range index {
		if len(ii) > 1 {
			locs := make([]Location, len(ii))
			for j, i := range ii {
				locs[j] = gg[i].Location()
			}
			sort.Sort(Locations(locs))

			force := ff[ii[0]].Key() == "source"
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
						gg[i] = WithLocation(gg[i], locs[j])
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
		gg[len(gg)-1] = nil
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
	if len(filters) == 0 {
		return TrueFilter
	}
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
	if len(filters) == 0 {
		return TrueFilter
	}
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
		return LocationWithin(f.Location(), lower, upper)
	}
}

// Overlap returns true if the location of the feature overlaps with the given
// bounds.
func Overlap(lower, upper int) Filter {
	return func(f Feature) bool {
		return LocationOverlap(f.Location(), lower, upper)
	}
}

// Key returns true if the key of a feature matches the given key string. If
// an empty string was given, the filter will always return true.
func Key(key string) Filter {
	if key == "" {
		return TrueFilter
	}
	return func(f Feature) bool { return f.Key() == key }
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
			for _, vv := range f.Values() {
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
			return f.Values().Has(name)
		}, nil
	}

	return func(f Feature) bool {
		if vv := f.Values().Get(name); vv != nil {
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
		props, err := toQualifier(head)
		if err != nil {
			return FalseFilter, err
		}
		filter = And(filter, props)
	}
	return filter, nil
}

// ForwardStrand returns true if the feature strictly resides on the forward
// strand.
func ForwardStrand(f Feature) bool {
	return CheckStrand(f.Location()) == StrandForward
}

// ReverseStrand returns true if the feature strictly resides on the reverse
// strand.
func ReverseStrand(f Feature) bool {
	return CheckStrand(f.Location()) == StrandReverse
}

// FeatureSlice represents a slice of Features.
type FeatureSlice []Feature

// Filter returns a FeatureSlice containing the features that match the given
// Filter within this FeatureSlice.
func (ff FeatureSlice) Filter(filter Filter) FeatureSlice {
	indices := make([]int, 0, len(ff))
	for i, f := range ff {
		if filter(NewFeature(f.Key(), f.Location(), f.Values())) {
			indices = append(indices, i)
		}
	}
	gg := make(FeatureSlice, len(indices))
	for i, index := range indices {
		gg[i] = ff[index]
	}
	return gg
}

// Len is the number of elements in the collection.
func (ff FeatureSlice) Len() int {
	return len(ff)
}

// Less reports whether the element with index i should sort before the element
// with index j.
func (ff FeatureSlice) Less(i, j int) bool {
	f, g := ff[i], ff[j]
	if f.Key() == "source" && g.Key() != "source" {
		return true
	}
	if f.Key() != "source" && g.Key() == "source" {
		return false
	}
	return LocationLess(f.Location(), g.Location())
}

// Swap the elements with indexes i and j.
func (ff FeatureSlice) Swap(i, j int) {
	ff[i], ff[j] = ff[j], ff[i]
}

// Insert takes the given Feature and inserts it into the sorted position in
// the FeatureSlice.
func (ff FeatureSlice) Insert(f Feature) FeatureSlice {
	i := 0
	for i < len(ff) && ff[i].Key() == "source" {
		i++
	}
	if f.Key() != "source" {
		i += sort.Search(len(ff[i:]), func(j int) bool {
			return LocationLess(f.Location(), ff[i+j].Location())
		})
	}

	ff = append(ff, nil)
	copy(ff[i+1:], ff[i:])
	ff[i] = f

	return ff
}
