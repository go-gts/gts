package gd

import "fmt"

type Pair struct {
	Key   string
	Value string
}

type OrderedDict struct {
	pairs []Pair
}

func NewOrderedDict() OrderedDict {
	return OrderedDict{pairs: make([]Pair, 0)}
}

func NewOrderedDictFromPairs(pairs []Pair) OrderedDict {
	return OrderedDict{pairs: pairs}
}

func (od OrderedDict) Len() int {
	return len(od.pairs)
}

func (od OrderedDict) Iter() []Pair {
  return od.pairs
}

func (od OrderedDict) indices(key string) []int {
	is := make([]int, 0)
	for i, pair := range od.pairs {
		if key == pair.Key {
			is = append(is, i)
		}
	}
	return is
}

func (od OrderedDict) Get(key string) []string {
	is := od.indices(key)
	ret := make([]string, len(is))
	for j, i := range is {
		ret[j] = od.pairs[i].Value
	}
	return ret
}

func (od OrderedDict) GetOne(key string) string {
	is := od.indices(key)
	if len(is) == 0 {
		panic(fmt.Errorf("OrderedDict does not have key: %s", key))
	}
	return od.pairs[is[0]].Value
}

func (od *OrderedDict) Set(key, value string) {
	od.pairs = append(od.pairs, Pair{Key: key, Value: value})
}
