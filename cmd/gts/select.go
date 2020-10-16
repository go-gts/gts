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
	flags.Register("select", "select features using the given feature selector(s)", selectFunc)
}

func selectFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	selectors := pos.Extra("selector", "feature selector (syntax: [feature_key][/[qualifier1][=regexp1]][/[qualifier2][=regexp2]]...)")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	invert := opt.Switch('v', "invert-match", "select features that do not match the given criteria")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	sort.Strings(*selectors)

	filters := make([]gts.Filter, len(*selectors))
	for i, selector := range *selectors {
		filter, err := gts.Selector(selector)
		if err != nil {
			return ctx.Raise(fmt.Errorf("invalid selector syntax: %v", err))
		}
		filters[i] = filter
	}
	filter := gts.Or(filters...)
	if *invert {
		filter = gts.Not(filter)
	}
	filter = gts.Or(gts.Key("source"), filter)

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
			{"selectors", *selectors},
			{"invert", *invert},
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
		ff := seq.Features().Filter(filter)
		seq = gts.WithFeatures(seq, ff)
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
