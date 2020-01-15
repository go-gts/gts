package main

import (
	"os"

	flags "gopkg.in/flags.v1"
)

func main() {
	name, desc := "gts", "the genome tool suite command line tool"
	os.Exit(flags.Run(name, desc, flags.Compile()))
}
