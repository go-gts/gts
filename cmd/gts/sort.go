package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

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
	h := newHash()
	pos, opt := flags.Flags()

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	reverse := opt.Switch('r', "reverse", "reverse the sort order")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	d, err := newIODelegate(h, *seqinPath, *seqoutPath)
	if err != nil {
		return ctx.Raise(err)
	}
	defer d.Close()

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"reverse", *reverse},
			{"filetype", filetype},
		})

		ok, err := d.Cache(data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}
	seqs := []gts.Sequence{}
	scanner := seqio.NewAutoScanner(d)
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

	w := bufio.NewWriter(d)

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
