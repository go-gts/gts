package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/internal/flags"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("insert", "insert a sequence into another sequence(s)", insertFunc)
}

func insertFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([selector|point|range][@modifier])")
	guestPath := pos.String("guest", "guest sequence file")

	var hostPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		hostPath = pos.String("host", "host sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	embed := opt.Switch('e', "embed", "extend existing feature locations when inserting instead of splitting them")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	insert := gts.Insert
	if *embed {
		insert = gts.Embed
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
	}

	hostFile := os.Stdin
	if hostPath != nil && *hostPath != "-" {
		f, err := os.Open(*hostPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *hostPath, err))
		}
		hostFile = f
		defer hostFile.Close()
	}

	guestFile, err := os.Open(*guestPath)
	if err != nil {
		return ctx.Raise(fmt.Errorf("failed to open file: %q: %v", *guestPath, err))
	}
	defer guestFile.Close()

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

	scanner := seqio.NewAutoScanner(guestFile)
	guests := []gts.Sequence{}
	for scanner.Scan() {
		guests = append(guests, scanner.Value())
	}
	if len(guests) == 0 {
		ctx.Raise(fmt.Errorf("guest sequence file %q does not contain a sequence", *guestPath))
	}

	w := bufio.NewWriter(seqoutFile)

	scanner = seqio.NewAutoScanner(hostFile)
	for scanner.Scan() {
		host := scanner.Value()

		rr := locate(host.Features())
		indices := make([]int, len(rr))
		for i, r := range rr {
			indices[i] = r.Head()
		}
		sort.Sort(sort.Reverse(sort.IntSlice(indices)))

		for _, guest := range guests {
			out := gts.Sequence(gts.Copy(host))
			for _, index := range indices {
				out = insert(out, index, guest)
			}
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
