package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("extract", "extract the sequences referenced by the features", extractFunc)
}

func containsRegion(rr []gts.Region, r gts.Region) bool {
	for i := range rr {
		if reflect.DeepEqual(rr[i], r) {
			return true
		}
	}
	return false
}

func extractFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	locstrs := pos.Extra("locator", "a locator string ([specifier][@modifier])")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	invert := opt.Switch('v', "invert-region", "extract the sequences that are not referenced by the features")

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

	if len(*locstrs) == 0 {
		*locstrs = append(*locstrs, "@^..$")
	}

	locators := make([]gts.Locator, len(*locstrs))

	for i, locstr := range *locstrs {
		locator, err := gts.AsLocator(locstr)
		if err != nil {
			return ctx.Raise(err)
		}
		locators[i] = locator
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"locators", *locstrs},
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

		rr := make([]gts.Region, 0)
		for _, locate := range locators {
			for _, r := range locate(seq) {
				if !containsRegion(rr, r) {
					rr = append(rr, r)
				}
			}
		}

		if *invert {
			// Support linear inversion only as topology is not well defined.
			rr = gts.InvertLinear(gts.Regions(rr), gts.Len(seq))
		}

		for _, region := range rr {
			if len(rr) == 1 || region.Len() != gts.Len(seq) {
				out := region.Locate(seq)
				if _, err := writer.WriteSeq(out); err != nil {
					return ctx.Raise(err)
				}
				if err := buffer.Flush(); err != nil {
					return ctx.Raise(err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}
