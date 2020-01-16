package gts

// Pair represents a pair of strings.
type Pair struct {
	Key   string `json:"key" yaml:"key" msgpack:"key"`
	Value string `json:"value" yaml:"value" msgpack:"value"`
}

// PairList represents a list of pairs.
type PairList []Pair

// Set the value for the given key.
func (pl *PairList) Set(key, value string) {
	for i, p := range *pl {
		if p.Key == key {
			(*pl)[i].Value = p.Value
			return
		}
	}
	*pl = append(*pl, Pair{key, value})
}
