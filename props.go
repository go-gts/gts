package gts

type Props [][]string

func (props Props) Index(key string) int {
	for i := range props {
		if props[i][0] == key {
			return i
		}
	}
	return -1
}

func (props Props) Has(name string) bool {
	return props.Index(name) >= 0
}

func (props Props) Keys() []string {
	keys := make([]string, len(props))
	for i := range props {
		keys[i] = props[i][0]
	}
	return keys
}

type Item struct {
	Key   string
	Value string
}

func (props Props) Items() []Item {
	items := make([]Item, 0)
	for _, key := range props.Keys() {
		for _, value := range props.Get(key) {
			items = append(items, Item{key, value})
		}
	}
	return items
}

func (props Props) Get(key string) []string {
	switch i := props.Index(key); i {
	case -1:
		return nil
	default:
		return props[i][1:]
	}
}

func (props *Props) Set(key string, values ...string) {
	prop := make([]string, len(values)+1)
	prop[0] = key
	copy(prop[1:], values)
	switch i := props.Index(key); i {
	case -1:
		*props = append(*props, prop)
	default:
		(*props)[i] = prop
	}
}

func (props *Props) Add(key string, values ...string) {
	switch i := props.Index(key); i {
	case -1:
		props.Set(key, values...)
	default:
		(*props)[i] = append((*props)[i], values...)
	}
}

func (props *Props) Del(key string) {
	if i := props.Index(key); i >= 0 {
		*props = append((*props)[:i], (*props)[i+1:]...)
	}
}
