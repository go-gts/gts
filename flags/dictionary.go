package flags

type Dictionary []Pair

func NewDictionary() *Dictionary {
	p := new([]Pair)
	*p = make([]Pair, 0)
	return (*Dictionary)(p)
}

func (d Dictionary) Index(key string) int {
	for i, p := range d {
		if p.Key == key {
			return i
		}
	}
	return -1
}

func (d Dictionary) Has(key string) bool {
	return d.Index(key) >= 0
}

func (d Dictionary) Get(key string) Value {
	if index := d.Index(key); index >= 0 {
		return d[index].Value
	}
	return nil
}

func (d *Dictionary) Set(key string, value Value) {
	if index := d.Index(key); index >= 0 {
		(*d)[index].Value = value
	}
	*d = append(*d, Pair{key, value})
}

func (d Dictionary) Len() int {
	return len([]Pair(d))
}

func (d Dictionary) Iter() []Pair {
	return []Pair(d)
}

type Subcommand struct {
	Func CommandFunc
	Desc string
}
