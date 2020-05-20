package gts

// Pair represents a pair of strings.
type Pair struct {
	Key   string
	Value string
}

// Dictionary represents a list of pairs.
type Dictionary []Pair

// Get the value for the given key.
func (d *Dictionary) Get(key string) []string {
	ret := []string{}
	for _, p := range *d {
		if p.Key == key {
			ret = append(ret, p.Value)
		}
	}
	return ret
}

// Set the value for the given key.
func (d *Dictionary) Set(key, value string) {
	for i, p := range *d {
		if p.Key == key {
			(*d)[i].Value = value
			return
		}
	}
	*d = append(*d, Pair{key, value})
}

// Del removes the value for the given key.
func (d *Dictionary) Del(key string) {
	for i, p := range *d {
		if p.Key == key {
			copy((*d)[i:], (*d)[i+1:])
			(*d)[len(*d)-1] = Pair{}
			(*d) = (*d)[:len(*d)-1]
		}
	}
}
