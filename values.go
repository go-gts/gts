package gts

// Values represents a collection of name-value list pairs.
type Values map[string][]string

// Get returns the values associated to the given name.
func (v Values) Get(key string) []string {
	if v == nil {
		return nil
	}
	if v, ok := v[key]; ok {
		return v
	}
	return nil
}

// Set overwrites the values associated to the given name.
func (v Values) Set(name string, values ...string) { v[name] = values }

// Add appends a value to the values associated to the given name.
func (v Values) Add(name, value string) { v[name] = append(v[name], value) }

// Del deletes the values associated to the given name.
func (v Values) Del(name string) { delete(v, name) }
