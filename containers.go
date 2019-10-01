package gd

import "fmt"

type Pair struct {
	Key   string
	Value string
}

type PairList struct {
	pairs []Pair
}

func NewPairList() PairList {
	return PairList{pairs: make([]Pair, 0)}
}

func NewPairListFromPairs(pairs []Pair) PairList {
	return PairList{pairs: pairs}
}

func (od PairList) Len() int {
	return len(od.pairs)
}

func (od PairList) Iter() []Pair {
	return od.pairs
}

func (od PairList) indices(key string) []int {
	is := make([]int, 0)
	for i, pair := range od.pairs {
		if key == pair.Key {
			is = append(is, i)
		}
	}
	return is
}

func (od PairList) Get(key string) []string {
	is := od.indices(key)
	ret := make([]string, len(is))
	for j, i := range is {
		ret[j] = od.pairs[i].Value
	}
	return ret
}

func (od PairList) GetOne(key string) string {
	is := od.indices(key)
	if len(is) == 0 {
		panic(fmt.Errorf("PairList does not have key: %s", key))
	}
	return od.pairs[is[0]].Value
}

func (od *PairList) Set(key, value string) {
	od.pairs = append(od.pairs, Pair{Key: key, Value: value})
}
