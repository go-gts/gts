package gts

import (
	"regexp"
	"sort"
	"strings"
)

// Feature represents a generic sequence feature. A Feature consists of three
// main values: a key representing the type of feature, the location of the
// feature, and its properties. The key should be restricted by the controlled
// vocabulary of the feature specification for the represented sequence format.
// Properties are represented as an ordered dictionary with multiple values for
// each key.
type Feature struct {
	Key   string
	Loc   Location
	Props Props
}

// NewFeature returns a new feature with the given arguments.
func NewFeature(key string, loc Location, props Props) Feature {
	return Feature{key, loc, props}
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

// Features represents a slice of Features.
type Features []Feature

// Filter returns a FeatureSlice containing the features that match the given
// Filter within this FeatureSlice.
func (ff Features) Filter(filter Filter) Features {
	indices := make([]int, 0, len(ff))
	for i, f := range ff {
		if filter(NewFeature(f.Key, f.Loc, f.Props)) {
			indices = append(indices, i)
		}
	}
	gg := make(Features, len(indices))
	for i, index := range indices {
		gg[i] = ff[index]
	}
	return gg
}

// Len is the number of elements in the collection.
func (ff Features) Len() int {
	return len(ff)
}

// Less reports whether the element with index i should sort before the element
// with index j.
func (ff Features) Less(i, j int) bool {
	return LocationLess(ff[i].Loc, ff[j].Loc)
}

// Swap the elements with indexes i and j.
func (ff Features) Swap(i, j int) {
	ff[i], ff[j] = ff[j], ff[i]
}

// Insert takes the given Feature and inserts it into the sorted position in
// the FeatureSlice.
func (ff Features) Insert(f Feature) Features {
	i := sort.Search(len(ff), func(i int) bool {
		return LocationLess(f.Loc, ff[i].Loc)
	})

	ff = append(ff, Feature{})
	copy(ff[i+1:], ff[i:])
	ff[i] = f

	return ff
}
