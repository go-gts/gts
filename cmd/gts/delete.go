package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-flip/flip"
	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("delete", "delete a region of the given sequence(s)", deleteFunc)
}

func deleteFunc(ctx *flags.Context) error {
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
	erase := opt.Switch('e', "erase", "remove features contained in the deleted regions")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
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

	delete := gts.Delete
	if *erase {
		delete = gts.Erase
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"locator", *locstr},
			{"erase", *erase},
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

		ss := gts.Minimize(locate(seq))
		flip.Flip(gts.BySegment(ss))
		for _, s := range ss {
			i, n := s.Head(), s.Len()
			seq = delete(seq, i, n)
		}

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
