package gts

// FeatureSelector represents a filter for selecting features.
type FeatureSelector func(f Feature) bool

// Any returns true for any feature matched.
func Any(f Feature) bool { return true }

// Key returns true if the key of a feature matches the given key.
func Key(key string) FeatureSelector {
	return func(f Feature) bool { return f.Key == key }
}

// And returns true if all selectors return true.
func And(ss ...FeatureSelector) FeatureSelector {
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
func Or(ss ...FeatureSelector) FeatureSelector {
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
func Not(s FeatureSelector) FeatureSelector {
	return func(f Feature) bool { return !s(f) }
}

// FeatureTable represents a feature table.
type FeatureTable interface {
	Select(sel FeatureSelector) []Feature
	Add(f Feature)
}
