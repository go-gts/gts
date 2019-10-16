package main

import (
	"errors"
	"fmt"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/gt1/flags"
	"github.com/ktnyt/gt1/seqio"
)

func init() {
	register("seq", "sequence manipulation commands", sequenceFunc)
}

func sequenceFragmentFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input sequence file")
	outfile := command.Outfile("output sequence file")
	window := command.Positional.Int("window", "fragmentation window size")
	step := command.Int('s', "step", 0, "fragmentation step size: defaults to window size")

	return command.Run(args, func() error {
		if *step <= 0 {
			*step = *window
		}

		scanner := seqio.NewScanner(infile)
		if !scanner.Scan() {
			return errors.New("input file cannot be interpreted as a sequence")
		}

		seq := scanner.Seq()

		fragments := gt1.Fragment(seq, *window, *step)

		pos := 0
		for _, fragment := range fragments {
			desc := fmt.Sprintf("fragment%d-%d", pos, pos+fragment.Len())
			fmt.Fprintln(outfile, seqio.FormatFasta(seqio.NewFasta(desc, fragment)))
			pos += fragment.Len()
		}

		return nil
	})
}

func sequenceSkewFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input sequence file")
	outfile := command.Outfile("output text file")
	setA := command.String('a', "set-a", "g", "character set to calculate skew (negative)")
	setB := command.String('b', "set-b", "c", "character set to calculate skew (positive)")
	cumulative := command.Switch('c', "cumulative", "calculate cumulative skew")

	return command.Run(args, func() error {
		scanner := seqio.NewScanner(infile)
		skew := 0.

		for scanner.Scan() {
			seq := scanner.Seq()
			if !*cumulative {
				skew = 0.
			}
			skew += gt1.Skew(seq, *setA, *setB)
			fmt.Fprintln(outfile, skew)
		}

		return nil
	})
}

func sequenceFunc(command *flags.Command, args []string) error {
	command.Command("fragment", "split input sequence into equal sized fragments", sequenceFragmentFunc)
	command.Command("skew", "calculate the skewness of the given sequence", sequenceSkewFunc)

	return command.Run(args)
}
