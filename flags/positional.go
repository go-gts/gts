package flags

import "fmt"

type Positional struct {
	Order  []string
	Values map[string]Value
	Usages map[string]string
}

func NewPositional() *Positional {
	return &Positional{
		Order:  make([]string, 0),
		Values: make(map[string]Value),
		Usages: make(map[string]string),
	}
}

func (positional Positional) hasValue(name string) bool {
	_, ok := positional.Values[name]
	return ok
}

func (positional Positional) Len() int {
	return len(positional.Order)
}

func (positional *Positional) Register(name string, value Value, usage string) {
	if positional.hasValue(name) {
		panic(fmt.Errorf("positional argument with name `%s` already exists", name))
	}

	positional.Order = append(positional.Order, name)
	positional.Values[name] = value
	positional.Usages[name] = usage
}

func (positional *Positional) Bool(name, usage string) *bool {
	value := NewBoolValue(false)
	positional.Register(name, value, usage)
	return (*bool)(value)
}

func (positional *Positional) Int(name, usage string) *int {
	value := NewIntValue(0)
	positional.Register(name, value, usage)
	return (*int)(value)
}

func (positional *Positional) String(name, usage string) *string {
	value := NewStringValue("")
	positional.Register(name, value, usage)
	return (*string)(value)
}

func (positional *Positional) Choice(name, usage string, choices ...string) *int {
	value := NewChoiceValue(choices, 0)
	positional.Register(name, value, usage)
	return value.Chosen
}

func (positional *Positional) Handle(args []string) ([]string, error) {
	arg := ""
	for _, name := range positional.Order {
		if len(args) == 0 {
			return nil, fmt.Errorf("missing mandatory argument `%s`", name)
		}
		arg, args = shift(args)
		if err := positional.Values[name].Set(arg); err != nil {
			return nil, fmt.Errorf("%s for mandatory argument `%s`", err, name)
		}
	}
	return args, nil
}
