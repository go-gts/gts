package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ktnyt/gd"
	"github.com/ktnyt/pars"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	log.Println("start")
	state := pars.NewState(file)
	ret, err := pars.Apply(gd.GenBankParser, state)
	if err != nil {
  	fmt.Println(string(state.Buffer[state.Index:]))
		log.Fatal(err)
	}
	log.Println("finish")

	gb := ret.(gd.GenBank)
	fmt.Println(gb.Format())
}
