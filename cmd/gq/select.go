package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/internal/flags"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("select", "select features using the given feature selector(s)", selectFunc)
}

type SelectProgram struct {
}

func selectFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	selectors := pos.Extra("selector", "feature selector (syntax: [feature_key][/[qualifier1][=regexp1]][/[qualifier2][=regexp2]]...)")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	// nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	// seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	strand := opt.String('s', "strand", "both", "strand to select features from (`both`, `forward`, or `reverse`)")
	invert := opt.Switch('v', "invert-match", "select features that do not match the given criteria")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	sort.Strings(*selectors)

	filters := make([]gts.Filter, len(*selectors))
	for i, selector := range *selectors {
		f, err := gts.Selector(selector)
		if err != nil {
			return ctx.Raise(fmt.Errorf("invalid selector syntax: %v", err))
		}
		filters[i] = f
	}
	filter := gts.Or(filters...)
	if *invert {
		filter = gts.Not(filter)
	}
	filter = gts.Or(gts.Key("source"), filter)

	switch *strand {
	case "forward":
		filter = gts.And(filter, gts.ForwardStrand)
	case "reverse":
		filter = gts.And(filter, gts.ReverseStrand)
	}

	/*
		if !*nocache {
			data := encodePayload([]tuple{
				{"command", strings.Join(ctx.Name, "-")},
				{"version", gts.Version.String()},
				{"selectors", *selectors},
				{"strand", *strand},
				{"invert", *invert},
				{"filetype", filetype},
			})
	*/

	var stream seqio.IOStream

	stream.ForEach(func(i int, header interface{}, ff gts.Features) (seqio.Callback, error) {
		if err := stream.PushHeader(header); err != nil {
			return nil, err
		}

		if err := stream.PushFeatures(ff.Filter(filter)); err != nil {
			return nil, err
		}

		return func(seq gts.Sequence) error {
			return stream.PushSequence(seq)
		}, nil
	})

	return nil
}
