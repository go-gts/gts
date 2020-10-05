package flags

import (
	"context"
	"fmt"
	"os"
)

// Flags returns a fresh pair of positional and optional argument sets.
func Flags() (*Positional, *Optional) {
	return newPositional(), newOptional()
}

var main = CommandSet{}

// Register a given command to the main CommandSet.
func Register(name, desc string, f Function) {
	main.Register(name, desc, f)
}

// Compile the main CommandSet.
func Compile() Function {
	return main.Compile()
}

// Run the given Function.
func Run(name, desc string, version Version, f Function) int {
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "--version" {
			fmt.Fprintln(os.Stdout, version)
			return 0
		}
	}
	ctx := &Context{[]string{name}, desc, args, context.Background()}
	if err := f(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
