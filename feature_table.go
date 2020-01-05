package gts

// FeatureFilter represents a filter for selecting features.
type FeatureFilter func(f Feature) bool

// Any returns true for any feature matched.
func Any(f Feature) bool { return true }

// And returns true if all selectors return true.
func And(ss ...FeatureFilter) FeatureFilter {
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
func Or(ss ...FeatureFilter) FeatureFilter {
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
func Not(s FeatureFilter) FeatureFilter {
	return func(f Feature) bool { return !s(f) }
}

// Key returns true if the key of a feature matches the given key.
func Key(key string) FeatureFilter {
	return func(f Feature) bool { return f.Key == key }
}

// FeatureTable represents a feature table.
type FeatureTable interface {
	Filter(ss ...FeatureFilter) []Feature
	Add(f Feature)
}
