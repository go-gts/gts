package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("extract", "extract the sequences referenced by the features", extractFunc)
}

func extractFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	modstr := opt.String('m', "range", "^..$", "location range modifier")

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

	mod, err := gts.AsModifier(*modstr)
	if err != nil {
		return ctx.Raise(fmt.Errorf("bad range modifier: %v", err))
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"range", mod.String()},
			{"filetype", filetype},
		})

		ok, err := d.Cache(data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	w := bufio.NewWriter(d)

	scanner := seqio.NewAutoScanner(d)
	for scanner.Scan() {
		seq := scanner.Value()
		ff := seq.Features().Filter(gts.Not(gts.Key("source")))
		for _, f := range ff {
			region := f.Location.Region()
			region = region.Resize(mod)
			out := region.Locate(seq)
			formatter := seqio.NewFormatter(out, filetype)
			if _, err := formatter.WriteTo(w); err != nil {
				return ctx.Raise(err)
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
