package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/cmd/cache"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("select", "select features using the given feature selector(s)", selectFunc)
}

func selectFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	selectors := pos.Extra("selector", "feature selector (syntax: [feature_key][/[qualifier1][=regexp1]][/[qualifier2][=regexp2]]...)")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
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

	cfg := cmd.NewIOConfig("gq", *seqinPath, *seqoutPath)
	if !*nocache {
		cfg = cfg.WithPayload(cache.Payload{
			"command":   strings.Join(ctx.Name, "-"),
			"version":   gts.Version.String(),
			"selectors": *selectors,
			"strand":    *strand,
			"invert":    *invert,
		})
	}

	r, w, err := cmd.HandleIO(cfg)
	if errors.Is(err, cmd.ErrCacheUsed) {
		return nil
	}
	if err != nil {
		return ctx.Raise(err)
	}

	istream, ostream, err := seqio.NewSeqIO(r, w)
	if err != nil {
		return ctx.Raise(err)
	}

	return ctx.Raise(
		istream.ForEach(
			func(i int, header interface{}, ff gts.Features) (seqio.SequenceHandler, error) {
				if err := ostream.PushHeader(header); err != nil {
					return nil, err
				}

				if err := ostream.PushFeatures(ff.Filter(filter)); err != nil {
					return nil, err
				}

				return func(seq gts.Sequence) error {
					return ostream.PushSequence(seq)
				}, nil
			},
		),
	)
}
