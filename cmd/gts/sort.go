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
	flags.Register("sort", "sort the list of sequences", sortFunc)
}

type byLength []gts.Sequence

func (ss byLength) Len() int {
	return len(ss)
}

func (ss byLength) Less(i, j int) bool {
	return gts.Len(ss[j]) < gts.Len(ss[i])
}

func (ss byLength) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

func sortFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	reverse := opt.Switch('r', "reverse", "reverse the sort order")
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

	seqs := []gts.Sequence{}

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		seqs = append(seqs, seq)
	}

	var iface sort.Interface
	iface = byLength(seqs)
	if *reverse {
		iface = sort.Reverse(iface)
	}
	sort.Sort(iface)

	for _, seq := range seqs {
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(w); err != nil {
			return ctx.Raise(err)
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
