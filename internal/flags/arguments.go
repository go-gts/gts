package flags

import (
	"fmt"
)

// Arguments is a map of names and values.
type Arguments map[string]Value

// Has tests if the given name is registered as an argument.
func (args Arguments) Has(name string) bool {
	_, ok := args[name]
	return ok
}

var shortNames = []rune("aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ")

type optionalName struct {
	Short rune
	Long  string
}

const intMax = ^int(0)

func runeLess(a, b rune) bool {
	x, y := intMax, intMax

	for i, r := range shortNames {
		if a == r {
			x = i
		}

		if b == r {
			y = i
		}
	}

	if x == intMax && y == intMax {
		return a < b
	}

	return x < y
}

type byShort []optionalName

func (names byShort) Len() int { return len(names) }

func (names byShort) Less(i, j int) bool {
	a, b := names[i], names[j]

	switch {
	case a.Short != 0 && b.Short != 0:
		return runeLess(a.Short, b.Short)

	case a.Short != 0:
		return runeLess(a.Short, []rune(b.Long)[0])

	case b.Short != 0:
		return runeLess([]rune(a.Long)[0], b.Short)

	default:
		return a.Long < b.Long
	}
}

func (names byShort) Swap(i, j int) {
	names[i], names[j] = names[j], names[i]
}

// Optional represents the optional command line arguments.
type Optional struct {
	Args  Arguments
	Alias map[rune]string
}

func newOptional() *Optional {
	return &Optional{make(Arguments), make(map[rune]string)}
}

func (opt *Optional) register(short rune, long string, value Value) {
	if opt.Args.Has(long) {
		panic(fmt.Errorf("optional argument with long name %q already exists", long))
	}

	if name, ok := opt.Alias[short]; ok {
		panic(fmt.Errorf("optional argument with short name `%c` already exists for name %q", short, name))
	}

	if short != 0 {
		opt.Alias[short] = long
	}

	opt.Args[long] = value
}

// Switch adds a command line switch to the optional argument list.
func (opt *Optional) Switch(short rune, long string) *bool {
	value := NewBoolValue(false)
	opt.register(short, long, value)
	return (*bool)(value)
}

// Int adds an integer flag to the optional argument list.
func (opt *Optional) Int(short rune, long string, init int) *int {
	value := NewIntValue(init)
	opt.register(short, long, value)
	return (*int)(value)
}

// IntSlice adds an integer slice flag to the optional argument list.
func (opt *Optional) IntSlice(short rune, long string, init []int) *[]int {
	value := NewIntSliceValue(init)
	opt.register(short, long, value)
	return (*[]int)(value)
}

// Float adds an float flag to the optional argument list.
func (opt *Optional) Float(short rune, long string, init float64) *float64 {
	value := NewFloatValue(init)
	opt.register(short, long, value)
	return (*float64)(value)
}

// FloatSlice adds an float slice flag to the optional argument list.
func (opt *Optional) FloatSlice(short rune, long string, init []float64) *[]float64 {
	value := NewFloatSliceValue(init)
	opt.register(short, long, value)
	return (*[]float64)(value)
}

// String adds an string flag to the optional argument list.
func (opt *Optional) String(short rune, long string, init string) *string {
	value := NewStringValue(init)
	opt.register(short, long, value)
	return (*string)(value)
}

// StringSlice adds an string slice flag to the optional argument list.
func (opt *Optional) StringSlice(short rune, long string, init []string) *[]string {
	value := NewStringSliceValue(init)
	opt.register(short, long, value)
	return (*[]string)(value)
}

// Positional represents the positional command line arguments.
type Positional struct {
	Order []string
	Args  Arguments
}

func newPositional() *Positional {
	return &Positional{make([]string, 0), Arguments{}}
}

func (pos *Positional) register(name string, value Value) {
	if pos.Args.Has(name) {
		panic(fmt.Errorf("positional argument with name %q already exists", name))
	}
	pos.Order = append(pos.Order, name)
	pos.Args[name] = value
}

// Len returns the number of positional arguments.
func (pos *Positional) Len() int {
	n := 0
	for _, value := range pos.Args {
		if _, ok := value.(*StringSliceValue); !ok {
			n++
		}
	}
	return n
}

// Switch adds a boolean switch to the positional argument list.
func (pos *Positional) Switch(name string) *bool {
	value := NewBoolValue(false)
	pos.register(name, value)
	return (*bool)(value)
}

// Int adds an int value to the positonal argument list.
func (pos *Positional) Int(name string) *int {
	value := NewIntValue(0)
	pos.register(name, value)
	return (*int)(value)
}

// Float adds a float value to the positional argument list.
func (pos *Positional) Float(name string) *float64 {
	value := NewFloatValue(0)
	pos.register(name, value)
	return (*float64)(value)
}

// String adds a string value to the positional argument list.
func (pos *Positional) String(name string) *string {
	value := NewStringValue("")
	pos.register(name, value)
	return (*string)(value)
}

// HasExtra returns true if extra arguments are available.
func (pos *Positional) HasExtra() bool {
	for _, value := range pos.Args {
		if _, ok := value.(*StringSliceValue); ok {
			return true
		}
	}
	return false
}

// Extra allows extra string values to be given.
func (pos *Positional) Extra(name string) *[]string {
	for key, value := range pos.Args {
		if _, ok := value.(*StringSliceValue); ok {
			panic(fmt.Errorf("extra arguments defined with name %q", key))
		}
	}

	value := NewStringSliceValue(nil)
	pos.register(name, value)
	return (*[]string)(value)
}
