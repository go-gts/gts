package gts

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Feature struct {
	Key   string
	Loc   Location
	Props Props
}

func NewFeature(key string, loc Location, props Props) Feature {
	return Feature{key, loc, props}
}

// Repair attempts to reconstruct features by joining features with identical
// feature keys and values which have adjacent locations.
func Repair(ff []Feature) []Feature {
	gg := make([]Feature, len(ff))
	copy(gg, ff)

	// Identify the features with similar keys and values.
	index := make(map[string][]int)
	for i, f := range gg {
		key := fmt.Sprintf("%s:%v", f.Key, f.Props)
		index[key] = append(index[key], i)
	}

	keep := make([]int, 0, len(gg))
	for _, indices := range index {
		if len(indices) > 0 {
			locs := make([]Location, len(indices))
			for j, i := range indices {
				locs[j] = gg[i].Loc
			}
			sort.Sort(Locations(locs))

			force := ff[indices[0]].Key == "source"
			list := LocationList{}
			for _, loc := range locs {
				list.Push(loc, force)
			}

			// DISCUSS: Should we join these locations?
			locs = list.Slice()

			// Some locations were merged.
			if len(locs) < len(indices) {
				for i, loc := range locs {
					gg[indices[i]].Loc = loc
				}
			}
			keep = append(keep, indices[:len(locs)]...)
		}
	}

	sort.Sort(sort.IntSlice(keep))

	i := 0
	for _, j := range keep {
		gg[i] = gg[j]
		i++
	}
	gg = gg[:len(keep)]

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
		return LocationWithin(f.Loc, lower, upper)
	}
}

// Overlap returns true if the location of the feature overlaps with the given
// bounds.
func Overlap(lower, upper int) Filter {
	return func(f Feature) bool {
		return LocationOverlap(f.Loc, lower, upper)
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
			for _, vv := range f.Props {
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
			return f.Props.Has(name)
		}, nil
	}

	return func(f Feature) bool {
		if vv := f.Props.Get(name); vv != nil {
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
	return CheckStrand(f.Loc) == StrandForward
}

// ReverseStrand returns true if the feature strictly resides on the reverse
// strand.
func ReverseStrand(f Feature) bool {
	return CheckStrand(f.Loc) == StrandReverse
}

// FeatureSlice represents a slice of Features.
type FeatureSlice []Feature

// Filter returns a FeatureSlice containing the features that match the given
// Filter within this FeatureSlice.
func (ff FeatureSlice) Filter(filter Filter) FeatureSlice {
	indices := make([]int, 0, len(ff))
	for i, f := range ff {
		if filter(NewFeature(f.Key, f.Loc, f.Props)) {
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
	if f.Key == "source" && g.Key != "source" {
		return true
	}
	if f.Key != "source" && g.Key == "source" {
		return false
	}
	return LocationLess(f.Loc, g.Loc)
}

// Swap the elements with indexes i and j.
func (ff FeatureSlice) Swap(i, j int) {
	ff[i], ff[j] = ff[j], ff[i]
}

// Insert takes the given Feature and inserts it into the sorted position in
// the FeatureSlice.
func (ff FeatureSlice) Insert(f Feature) FeatureSlice {
	i := 0
	for i < len(ff) && ff[i].Key == "source" {
		i++
	}
	if f.Key != "source" {
		i += sort.Search(len(ff[i:]), func(j int) bool {
			return LocationLess(f.Loc, ff[i+j].Loc)
		})
	}

	ff = append(ff, Feature{})
	copy(ff[i+1:], ff[i:])
	ff[i] = f

	return ff
}
