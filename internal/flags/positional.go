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

// Len returns the number of positional arguments.
func (pos *Positional) Len() int {
	n := 0
	for _, arg := range pos.Args {
		if _, ok := arg.Value.(*StringSliceValue); !ok {
			n++
		}
	}
	return n
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

// HasExtra returns true if extra arguments are available.
func (pos *Positional) HasExtra() bool {
	for _, arg := range pos.Args {
		if _, ok := arg.Value.(*StringSliceValue); ok {
			return true
		}
	}
	return false
}

// Extra allows extra string values to be given.
func (pos *Positional) Extra(name, usage string) *[]string {
	for key, arg := range pos.Args {
		if _, ok := arg.Value.(*StringSliceValue); ok {
			panic(fmt.Errorf("extra arguments defined with name %q", key))
		}
	}
	value := NewStringSliceValue(nil)
	pos.register(name, value, usage)
	return (*[]string)(value)
}
