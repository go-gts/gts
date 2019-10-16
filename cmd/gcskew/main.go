package main

import (
	"fmt"
	"os"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/gt1/flags"
	"github.com/ktnyt/gt1/seqio"
)

func main() {
	prog, args := os.Args[0], os.Args[1:]
	command := flags.NewCommand(prog, "Calculate the GC skew (or related metrics) of the given sequence.")
	infile := command.Infile("input sequence file")
	outfile := command.Outfile("output text file")
	cumulative := command.Switch('c', "cumulative", "calculate cumulative skew")
	metric := command.Choice('m', "metric", "bases to calculate skewness for", "gc", "at", "purine", "keto")
	window := command.Int('w', "window", 10000, "window size")
	step := command.Int('s', "step", *window, "step size")

	sets := [][]string{
		[]string{"g", "c"},
		[]string{"a", "t"},
		[]string{"ag", "tc"},
		[]string{"tg", "ac"},
	}

	if err := command.Run(args, func() error {
		scanner := seqio.NewScanner(infile)
		if !scanner.Scan() {
			return fmt.Errorf("input file could not be interpreted as a sequence")
		}

		seq := scanner.Seq()

		set := sets[*metric]

		fragments := gt1.Fragment(seq, *window, *step)
		pos := 0
		skew := 0.

		names := []string{"GC Skew", "AT Skew", "Purine Skew", "Keto Skew"}

		fmt.Fprintf(outfile, "Position\t%s\n", names[*metric])

		for _, fragment := range fragments {
			if !*cumulative {
				skew = 0.
			}
			skew += gt1.Skew(fragment, set[0], set[1])
			fmt.Fprintf(outfile, "%d\t%f\n", pos, skew)
			pos += *step
		}

		return nil
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
