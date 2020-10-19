package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("join", "join the sequences contained in the files", joinFunc)
}

func joinFunc(ctx *flags.Context) error {
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
	circular := opt.Switch('c', "circular", "output the sequence as circular if possible")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	d, err := newIODelegate(*seqinPath, *seqoutPath)
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
			{"circular", *circular},
			{"filetype", filetype},
		})

		ok, err := d.TryCache(h, data)
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

	seq := gts.Concat(seqs...)

	if *circular {
		seq = gts.WithTopology(seq, gts.Circular)
	}

	formatter := seqio.NewFormatter(seq, filetype)
	if _, err := formatter.WriteTo(d); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}
