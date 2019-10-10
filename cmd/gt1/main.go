package main

import (
	"fmt"
	"os"

	"github.com/ktnyt/gt1/flags"
)

var command *flags.Command

func register(name, desc string, cmd flags.CommandFunc) {
	if command == nil {
		command = flags.NewCommand(os.Args[0], "genome toolkit version 1")
	}
	command.Command(name, desc, cmd)
}

func main() {
	if err := command.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
