package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/flags"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("complement", "compute the complement of the given sequence", sequenceComplement)
	flags.Register("search", "search for a subsequence and annotate its results", sequenceSearch)
}

func sequenceComplement(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

	seqoutFile := os.Stdout
	if *seqoutPath != "-" {
		f, err := os.Create(*seqoutPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *seqoutPath, err))
		}
		seqoutFile = f
		defer seqoutFile.Close()
	}

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	w := bufio.NewWriter(seqoutFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		seq = gts.Complement(seq)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(w); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

func sequenceSearch(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	queryString := pos.String("query", "query sequence (will attempt to open as file)")

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	featureKey := opt.String('k', "key", "misc_feature", "key for the reported oligomer region features")
	qfstrs := opt.StringSlice('q', "qualifier", nil, "qualifier key-value pairs (syntax: key=value))")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	qfs := gts.Values{}
	for _, s := range *qfstrs {
		switch i := strings.IndexByte(s, '='); i {
		case -1:
			qfs.Add(s, "")
		default:
			qfs.Add(s[:i], s[i+1:])
		}
	}

	queries := []gts.Sequence{}

	queryFile, err := os.Open(*queryString)
	if err == nil {
		scanner := seqio.NewAutoScanner(queryFile)
		for scanner.Scan() {
			queries = append(queries, scanner.Value())
		}
	} else {
		queries = append(queries, gts.New(nil, nil, []byte(*queryString)))
	}

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

	seqoutFile := os.Stdout
	if *seqoutPath != "-" {
		f, err := os.Create(*seqoutPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *seqoutPath, err))
		}
		seqoutFile = f
		defer seqoutFile.Close()
	}

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	w := bufio.NewWriter(seqoutFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		ff := seq.Features()
		for _, query := range queries {
			locs := gts.Search(seq, query)
			for _, loc := range locs {
				ff = ff.Insert(gts.Feature{
					Key:        *featureKey,
					Location:   loc,
					Qualifiers: qfs,
				})
			}
		}
		seq = gts.WithFeatures(seq, ff)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(w); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}
