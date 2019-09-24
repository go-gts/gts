package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ktnyt/gd"
	"github.com/ktnyt/pars"
)

var seqin = flag.String("seqin", "", "Sequence(s) to read")
var seqout = flag.String("seqout", "", "Outfile")

func main() {
	flag.Parse()

	file, err := os.Open(*seqin)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	log.Println("start")

	start := time.Now()

	state := pars.NewState(file)
	result := &pars.Result{}
	if err := gd.GenBankParser(state, result); err != nil {
		fmt.Println(string(state.Buffer[state.Index:]))
		log.Fatal(err)
	}

	elapsed := time.Since(start)

	log.Println("finished in", elapsed)

	gb := result.Value.(gd.GenBank)
	f, err := os.Create(*seqout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(f, "%s", gb.Format())
}
