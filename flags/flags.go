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
func Run(name, desc string, f Function) int {
	ctx := &Context{name, desc, os.Args[1:], context.Background()}
	if err := f(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
