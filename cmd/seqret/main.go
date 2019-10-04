package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

var seqin = flag.String("seqin", "", "Sequence(s) to read")
var seqout = flag.String("seqout", "", "Outfile")

func read(filename string) gt1.Record {
	file, err := os.Open(*seqin)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	log.Println("start read")
	start := time.Now()

	state := pars.NewState(file)
	result := &pars.Result{}
	if err := gt1.GenBankParser(state, result); err != nil {
		fmt.Println(string(state.Buffer[state.Index:]))
		log.Fatal(err)
	}
	gb := result.Value.(gt1.Record)

	elapsed := time.Since(start)
	log.Println("finished in", elapsed)

	return gb
}

func write(filename string, gb gt1.Record) {
	log.Println("start write")
	start := time.Now()

	f, err := os.Create(*seqout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(f, "%s", gt1.FormatGenBank(gb))

	elapsed := time.Since(start)

	log.Println("finished in", elapsed)
}

func main() {
	flag.Parse()
	write(*seqout, read(*seqin))
}
