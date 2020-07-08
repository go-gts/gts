package main

import (
	"os"

	"github.com/go-gts/gts/flags"
)

func main() {
	name, desc := "gts", "the genomics tool suite command line tool"
	version := flags.Version{
		Major: 0,
		Minor: 11,
		Patch: 9,
	}
	os.Exit(flags.Run(name, desc, version, flags.Compile()))
}
