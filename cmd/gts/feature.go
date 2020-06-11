package main

import (
	"fmt"
	"os"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/flags"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("clear", "remove all features from the sequence (excluding source features)", FeatureClear)
}

// FeatureClear will clear all features except for sources.
func FeatureClear(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var inpath *string

	if isTerminal(os.Stdin.Fd()) {
		inpath = pos.String("input", "file to input (may be omitted if standard input is provided)")
	}

	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	outpath := opt.String('o', "output", "-", "file to output (specifying `-` will force standard output)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	infile := os.Stdin
	if inpath != nil && *inpath != "-" {
		f, err := os.Open(*inpath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *inpath, err))
		}
		infile = f
	}

	outfile := os.Stdout
	if *outpath != "-" {
		f, err := os.Create(*outpath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *inpath, err))
		}
		outfile = f
	}

	filetype := seqio.Detect(*outpath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	scanner := seqio.NewAutoScanner(infile)
	for scanner.Scan() {
		seq := scanner.Value()
		ff := seq.Features().Filter(gts.Key("source"))
		seq = gts.WithFeatures(seq, ff)
		w := seqio.NewFormatter(seq, filetype)
		_, err := w.WriteTo(outfile)
		if err != nil {
			return ctx.Raise(err)
		}
	}

	return nil
}
