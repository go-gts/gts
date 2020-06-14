package main

import (
	"fmt"
	"os"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/flags"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("insert", "insert a sequence into another sequence", sequenceInsert)
}

func sequenceInsert(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	n := pos.Int("position", "a zero-indexed position to insert the guest sequence")
	guestPath := pos.String("guest", "guest sequence file")

	var hostPath *string
	if isTerminal(os.Stdin.Fd()) {
		hostPath = pos.String("host", "host sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
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
	if !scanner.Scan() {
		ctx.Raise(fmt.Errorf("guest sequence file %q does not contain a sequence", *guestPath))
	}
	guest := scanner.Value()

	scanner = seqio.NewAutoScanner(hostFile)
	for scanner.Scan() {
		host := scanner.Value()
		host = gts.Insert(host, *n, guest)
		w := seqio.NewFormatter(host, filetype)
		_, err := w.WriteTo(seqoutFile)
		if err != nil {
			return ctx.Raise(err)
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}
