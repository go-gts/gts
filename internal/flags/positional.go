package flags

import "fmt"

// Positional represents the positional command line arguments.
type Positional struct {
	Order []string
	Args  Arguments
}

func newPositional() *Positional {
	return &Positional{make([]string, 0), Arguments{}}
}

func (pos *Positional) register(name string, value Value, usage string) {
	if pos.Args.Has(name) {
		panic(fmt.Errorf("positional argument with name %q already exists", name))
	}
	pos.Order = append(pos.Order, name)
	pos.Args[name] = Argument{value, usage}
}

// Switch adds a boolean switch to the positional argument list.
func (pos *Positional) Switch(name, usage string) *bool {
	value := NewBoolValue(false)
	pos.register(name, value, usage)
	return (*bool)(value)
}

// Int adds an int value to the positonal argument list.
func (pos *Positional) Int(name, usage string) *int {
	value := NewIntValue(0)
	pos.register(name, value, usage)
	return (*int)(value)
}

// Float adds a float value to the positional argument list.
func (pos *Positional) Float(name, usage string) *float64 {
	value := NewFloatValue(0)
	pos.register(name, value, usage)
	return (*float64)(value)
}

// String adds a string value to the positional argument list.
func (pos *Positional) String(name, usage string) *string {
	value := NewStringValue("")
	pos.register(name, value, usage)
	return (*string)(value)
}
