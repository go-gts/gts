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
	flags.Register("define", "define a new feature", defineFunc)
}

func defineFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	key := pos.String("key", "feature key")
	locstr := pos.String("location", "feature location")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	qfstrs := opt.StringSlice('q', "qualifier", nil, "qualifier key-value pairs (syntax: key=value))")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	loc, err := gts.AsLocation(*locstr)
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

	order := make(map[string]int)
	qfs := gts.Values{}
	for _, s := range *qfstrs {
		name, value := s, ""
		if i := strings.IndexByte(s, '='); i >= 0 {
			name, value = s[:i], s[i+1:]
		}
		qfs.Add(name, value)
		order[name] = len(order)
	}

	f := gts.Feature{
		Key:        *key,
		Location:   loc,
		Qualifiers: qfs,
		Order:      order,
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"key", *key},
			{"location", loc.String()},
			{"qualifiers", *qfstrs},
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

		ff := seq.Features()
		ff = ff.Insert(f)
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
