package gts

import (
	"strings"
)

// Filter represents a filter for selecting features.
type Filter func(f Feature) bool

// And creates an And filter for this and the given filter.
func (f Filter) And(g Filter) Filter {
	return And(f, g)
}

// Or creates an Or filter for this and the given filter.
func (f Filter) Or(g Filter) Filter {
	return Or(f, g)
}

// Any returns true for any feature matched.
func Any(f Feature) bool { return true }

// And returns true if all selectors return true.
func And(ss ...Filter) Filter {
	return func(f Feature) bool {
		for _, s := range ss {
			if !s(f) {
				return false
			}
		}
		return true
	}
}

// Or returns true if any of the selectors return true.
func Or(ss ...Filter) Filter {
	return func(f Feature) bool {
		for _, s := range ss {
			if s(f) {
				return true
			}
		}
		return false
	}
}

// Not returns true if the given selector returns false.
func Not(s Filter) Filter {
	return func(f Feature) bool { return !s(f) }
}

// Key returns true if the key of a feature matches the given key.
func Key(key string) Filter {
	if key == "" {
		return func(f Feature) bool { return true }
	}
	return func(f Feature) bool { return f.Key == key }
}

// Qualifier returns true if the value for the given qualifier name matches
// the given expression.
func Qualifier(name string, query string) Filter {
	return func(f Feature) bool {
		if values, ok := f.Qualifiers[name]; ok {
			for _, value := range values {
				if strings.Contains(value, query) {
					return true
				}
			}
		}
		return false
	}
}

func asQualifier(s string) Filter {
	if i := strings.IndexByte(s, '='); i >= 0 {
		return Qualifier(s[:i], s[i+1:])
	}
	return Qualifier(s, "")
}

func selectorSplit(s string) (string, string) {
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

// ParseSelector creates a Filter from the given selector string.
func ParseSelector(s string) Filter {
	head, tail := selectorSplit(s)
	filter := Key(head)
	for tail != "" {
		head, tail = selectorSplit(tail)
		filter = filter.And(asQualifier(head))
	}
	return filter
}
