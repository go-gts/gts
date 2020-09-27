package main

import (
	"os"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/flags"
)

func main() {
	name, desc := "gts", "the genome transformation subprograms command line tool"
	os.Exit(flags.Run(name, desc, gts.Version, flags.Compile()))
}
