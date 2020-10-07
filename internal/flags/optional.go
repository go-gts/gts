package flags

import (
	"fmt"
)

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
	return &Optional{Arguments{}, make(map[rune]string)}
}

func (opt *Optional) register(short rune, long string, value Value, usage string) {
	if opt.Args.Has(long) {
		panic(fmt.Errorf("optional argument with long name %q already exists", long))
	}
	if name, ok := opt.Alias[short]; ok {
		panic(fmt.Errorf("optional argument with short name `%c` already exists for name %q", short, name))
	}
	if short != 0 {
		opt.Alias[short] = long
	}
	opt.Args[long] = Argument{value, usage}
}

// Switch adds a command line switch to the optional argument list.
func (opt *Optional) Switch(short rune, long string, usage string) *bool {
	value := NewBoolValue(false)
	opt.register(short, long, value, usage)
	return (*bool)(value)
}

// Int adds an integer flag to the optional argument list.
func (opt *Optional) Int(short rune, long string, init int, usage string) *int {
	value := NewIntValue(init)
	opt.register(short, long, value, usage)
	return (*int)(value)
}

// IntSlice adds an integer slice flag to the optional argument list.
func (opt *Optional) IntSlice(short rune, long string, init []int, usage string) *[]int {
	value := NewIntSliceValue(init)
	opt.register(short, long, value, usage)
	return (*[]int)(value)
}

// Float adds an float flag to the optional argument list.
func (opt *Optional) Float(short rune, long string, init float64, usage string) *float64 {
	value := NewFloatValue(init)
	opt.register(short, long, value, usage)
	return (*float64)(value)
}

// FloatSlice adds an float slice flag to the optional argument list.
func (opt *Optional) FloatSlice(short rune, long string, init []float64, usage string) *[]float64 {
	value := NewFloatSliceValue(init)
	opt.register(short, long, value, usage)
	return (*[]float64)(value)
}

// String adds an string flag to the optional argument list.
func (opt *Optional) String(short rune, long string, init string, usage string) *string {
	value := NewStringValue(init)
	opt.register(short, long, value, usage)
	return (*string)(value)
}

// StringSlice adds an string slice flag to the optional argument list.
func (opt *Optional) StringSlice(short rune, long string, init []string, usage string) *[]string {
	value := NewStringSliceValue(init)
	opt.register(short, long, value, usage)
	return (*[]string)(value)
}
