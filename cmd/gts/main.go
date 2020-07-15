package main

import (
	"os"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/flags"
)

func main() {
	name, desc := "gts", "the genomics tool suite command line tool"
	os.Exit(flags.Run(name, desc, gts.Version, flags.Compile()))
}
