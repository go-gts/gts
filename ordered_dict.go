package gd

type OrderedDict struct {
	keys []string
	dict map[string]string
}

func NewOrderedDict() OrderedDict {
	return OrderedDict{
		keys: make([]string, 0),
		dict: make(map[string]string),
	}
}

func (od OrderedDict) Len() int {
	return len(od.keys)
}

func (od OrderedDict) Keys() []string {
	return od.keys
}

func (od OrderedDict) Get(key string) string {
	return od.dict[key]
}

func (od *OrderedDict) Set(key, value string) {
	if _, ok := od.dict[key]; !ok {
		od.keys = append(od.keys, key)
	}
	od.dict[key] = value
}
