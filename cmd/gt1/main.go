package main

import (
	"fmt"
	"io"
	"os"

	termutil "github.com/andrew-d/go-termutil"
	"github.com/ktnyt/gt1/flags"
)

func Open(s string) *os.File {
	f, err := os.Open(s)
	if err != nil {
		panic(err)
	}
	return f
}

func Create(s string) *os.File {
	f, err := os.Create(s)
	if err != nil {
		panic(err)
	}
	return f
}

func getReaderAndWriter(infile, outfile string) (io.Reader, io.Writer) {
	r, w := os.Stdin, os.Stdout

	if infile != "" {
		if outfile != "" {
			r = Open(infile)
			w = Open(outfile)
		} else {
			if termutil.Isatty(os.Stdin.Fd()) {
				r = Open(infile)
			} else {
				w = Open(infile)
			}
		}
	}

	return r, w
}

var parser *flags.Parser

func register(name string, cmd flags.Command) {
	if parser == nil {
		parser = flags.NewParser(os.Args[0], version)
	}
	parser.Command(name, cmd)
}

func main() {
	args := os.Args[1:]

	_, err := parser.Parse(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
