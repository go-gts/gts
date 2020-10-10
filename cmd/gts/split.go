package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("split", "split the sequence at the provided locations", splitFunc)
}

func splitFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([selector|point|range][@modifier])")

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
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
		rr := locate(seq.Features())

		top := gts.Linear
		switch v := seq.(type) {
		case seqio.GenBank:
			top = v.Fields.Topology
		}

		switch {
		case len(rr) == 0:
			formatter := seqio.NewFormatter(seq, filetype)
			if _, err := formatter.WriteTo(w); err != nil {
				return ctx.Raise(err)
			}

		case len(rr) == 1 && top == gts.Circular:
			seq = gts.Rotate(seq, -rr.Head())
			seq = gts.WithTopology(seq, gts.Linear)
			formatter := seqio.NewFormatter(seq, filetype)
			if _, err := formatter.WriteTo(w); err != nil {
				return ctx.Raise(err)
			}

		default:
			unique := make(map[int]interface{})
			for _, r := range rr {
				unique[r.Head()] = nil
			}

			heads := make([]int, len(unique))
			i := 0
			for head := range unique {
				heads[i] = head
				i++
			}

			sort.Ints(heads)

			splits := make([]int, len(heads)+2)
			if top == gts.Circular {
				splits[0] = heads[len(heads)-1]
				splits = splits[:len(splits)-1]
			} else {
				splits[len(splits)-1] = gts.Len(seq)
			}
			for i, head := range heads {
				splits[i+1] = head
			}

			for i, tail := range splits[1:] {
				head := splits[i]
				sub := gts.Slice(seq, head, tail)
				sub = gts.WithTopology(sub, gts.Linear)
				formatter := seqio.NewFormatter(sub, filetype)
				if _, err := formatter.WriteTo(w); err != nil {
					return ctx.Raise(err)
				}
			}
		}

		if err := w.Flush(); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}
