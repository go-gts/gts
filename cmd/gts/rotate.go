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
	flags.Register("rotate", "shift the coordinates of a circular sequence", rotateFunc)
}

func rotateFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([modifier|selector|point|range][@modifier])")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
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
			{"locator", *locstr},
			{"filetype", filetype},
		})

		ok, err := d.TryCache(h, data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	scanner := seqio.NewAutoScanner(d)
	buffer := bufio.NewWriter(d)
	writer := seqio.NewWriter(buffer, filetype)

	for scanner.Scan() {
		seq := scanner.Value()
		rr := locate(seq)
		if len(rr) > 0 {
			seq = gts.Rotate(seq, -rr[0].Head())
		}
		seq = gts.WithTopology(seq, gts.Circular)
		if _, err := writer.WriteSeq(seq); err != nil {
			return ctx.Raise(err)
		}

		if err := buffer.Flush(); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}
