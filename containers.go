package gt1

import "fmt"

type Pair struct {
	Key   string
	Value string
}

type PairList struct {
	pairs []Pair
}

func NewPairList() *PairList {
	return &PairList{pairs: make([]Pair, 0)}
}

func NewPairListFromPairs(pairs []Pair) *PairList {
	return &PairList{pairs: pairs}
}

func (pl PairList) Len() int {
	return len(pl.pairs)
}

func (pl PairList) Iter() []Pair {
	return pl.pairs
}

func (pl PairList) indices(key string, not bool) []int {
	is := make([]int, 0)
	for i, pair := range pl.pairs {
		if !not && key == pair.Key {
			is = append(is, i)
		}
		if not && key != pair.Key {
			is = append(is, i)
		}
	}
	return is
}

func (pl PairList) Get(key string) string {
	is := pl.indices(key, false)
	if len(is) == 0 {
		panic(fmt.Errorf("PairList does not have key: %s", key))
	}
	return pl.pairs[is[0]].Value
}

func (pl PairList) All(key string) []string {
	is := pl.indices(key, false)
	ret := make([]string, len(is))
	for j, i := range is {
		ret[j] = pl.pairs[i].Value
	}
	return ret
}

func (pl *PairList) Add(key, value string) {
	pl.pairs = append(pl.pairs, Pair{Key: key, Value: value})
}

func (pl *PairList) Del(key string) {
	is := pl.indices(key, true)
	pairs := make([]Pair, len(is))
	for j, i := range is {
		pairs[j] = pl.pairs[i]
	}
	pl.pairs = pairs
}

func (pl *PairList) Set(key, value string) {
	pl.Del(key)
	pl.Add(key, value)
}
