package main

import (
	"os"

	"github.com/go-gts/gts/flags"
)

func main() {
	name, desc := "gts", "the genomics tool suite command line tool"
	version := flags.Version{
		Major: 0,
		Minor: 9,
		Patch: 2,
	}
	os.Exit(flags.Run(name, desc, version, flags.Compile()))
}
