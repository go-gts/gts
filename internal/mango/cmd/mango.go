package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-gts/gts/internal/mango"
	"github.com/goccy/go-yaml"
)

func run() error {
	infile, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}

	dec := yaml.NewDecoder(infile)
	cmd := mango.Command{}
	if err := dec.Decode(&cmd); err != nil {
		return err
	}

	outfile, err := os.Create(fmt.Sprintf("%s.1", cmd.Name))
	if err != nil {
		return err
	}

	if err := cmd.Roff(outfile); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
