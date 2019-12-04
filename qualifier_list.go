package gt1

// Qualifiers represents a collection of feature qualifiers.
type Qualifiers map[string][]string

// Get will return the qualifier values associated to the given name.
func (qs Qualifiers) Get(key string) []string {
	if qs == nil {
		return nil
	}
	if v, ok := qs[key]; ok {
		return v
	}
	return nil
}

// Set will overwrite the qualifier values associated to the given name.
func (qs Qualifiers) Set(name string, values ...string) { qs[name] = values }

// Add will add a value to the qualifier associated to the given name.
func (qs Qualifiers) Add(name, value string) { qs[name] = append(qs[name], value) }

// Del will delete the qualifier values associated to the given name.
func (qs Qualifiers) Del(name string) { delete(qs, name) }
