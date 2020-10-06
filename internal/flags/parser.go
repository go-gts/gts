package flags

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ArgumentType represents the type of argument.
type ArgumentType int

const (
	// LongType represents a long flag argument.
	LongType ArgumentType = iota

	// ShortType represents a short flag argument.
	ShortType

	// ValueType represents a plain value argument.
	ValueType

	// Terminator represents an argument list terminator `--`.
	Terminator
)

func mustMatchString(pattern string, s string) bool {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		panic(err)
	}
	return matched
}

// TypeOf returns the type of the given argument.
func TypeOf(s string) ArgumentType {
	if s == "--" {
		return Terminator
	}
	if strings.HasPrefix(s, "--") {
		return LongType
	}
	if mustMatchString("-[^0-9]+", s) {
		return ShortType
	}
	return ValueType
}

var errHelp = errors.New("help")
var errRonn = errors.New("ronn")
var errComp = errors.New("comp")

// Parse will parse the argument list according to the positional and optional
// argument lists provided and return extraneous argument elements and an error
// value if present.
func Parse(pos *Positional, opt *Optional, args []string) ([]string, error) {
	head := ""
	extra := []string{}
	terminated := false

	for len(args) > 0 && !terminated {
		head, args = shift(args)

		switch head {
		case "generate-ronn-templates":
			return nil, errRonn
		case "generate-completions":
			return nil, errComp
		}

		switch TypeOf(head) {
		case LongType:
			long := head[2:]

			if long == "help" {
				return nil, errHelp
			}

			switch i := strings.IndexByte(long, '='); i {
			case -1:
				arg, ok := opt.Args[long]
				if !ok {
					return nil, fmt.Errorf("unknown flag %q", long)
				}

				switch v := arg.Value.(type) {
				case *BoolValue:
					*v = BoolValue(true)
				case SliceValue:
					for len(args)+len(extra) > pos.Len() && TypeOf(args[0]) == ValueType {
						head, args = shift(args)
						if err := v.Set(head); err != nil {
							return nil, fmt.Errorf("while setting value for flag %q: %v", long, err)
						}
					}
				default:
					head, args = shift(args)
					if TypeOf(head) != ValueType {
						return nil, fmt.Errorf("while setting value for flag %q: no value given", long)
					}
					if err := v.Set(head); err != nil {
						return nil, fmt.Errorf("while setting value for flag %q: %v", long, err)
					}
				}

			default:
				name, value := long[:1], long[i+1:]
				arg, ok := opt.Args[name]
				if !ok {
					return nil, fmt.Errorf("unknown flag %q", name)
				}
				if err := arg.Value.Set(value); err != nil {
					return nil, fmt.Errorf("while setting value for flag %q: %v", long, err)
				}
			}

		case ShortType:
			rr := []rune(head[1:])
			var r rune

			for len(rr) > 0 {
				r, rr = rr[0], rr[1:]

				if r == 'h' {
					return nil, errHelp
				}

				name, ok := opt.Alias[r]
				if !ok {
					return nil, fmt.Errorf("unknown short option `%c`", r)
				}

				switch v := opt.Args[name].Value.(type) {
				case *BoolValue:
					*v = BoolValue(true)
				case SliceValue:
					for len(args)+len(extra) > pos.Len() && TypeOf(args[0]) == ValueType {
						head, args = shift(args)
						if err := v.Set(head); err != nil {
							return nil, fmt.Errorf("while setting value for flag %q: %v", name, err)
						}
					}
				default:
					head, args = shift(args)
					if TypeOf(head) != ValueType {
						return nil, fmt.Errorf("while setting value for flag %q: no value given", name)
					}
					if err := v.Set(head); err != nil {
						return nil, fmt.Errorf("while setting value for flag %q: %v", name, err)
					}
				}
			}

		case ValueType:
			extra = append(extra, head)
		case Terminator:
			extra = append(extra, args...)
			terminated = true
		}
	}

	n := 0
	for i, name := range pos.Order {
		if len(extra) == 0 {
			list := make([]string, len(pos.Order)-i)
			for j, name := range pos.Order[i:] {
				list[j] = fmt.Sprintf("%q", name)
			}
			missing := strings.Join(list, ", ")
			return extra, fmt.Errorf("missing positional arguments(s): %s", missing)
		}

		switch pos.Args[name].Value.(type) {
		case *StringSliceValue:
			for len(extra)+n > pos.Len() {
				head, extra = shift(extra)
				if err := pos.Args[name].Value.Set(head); err != nil {
					return extra, err
				}
			}
		default:
			head, extra = shift(extra)
			if err := pos.Args[name].Value.Set(head); err != nil {
				return extra, err
			}
			n++
		}
	}

	return extra, nil
}
