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
	flags.Register("search", "search for a subsequence and annotate its results", searchFunc)
}

func searchFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	queryPath := pos.String("query", "query sequence file (will be interpreted literally if preceded with @)")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	featureKey := opt.String('k', "key", "misc_feature", "key for the reported oligomer region features")
	propstrs := opt.StringSlice('q', "qualifier", nil, "qualifier key-value pairs (syntax: key=value))")
	exact := opt.Switch('e', "exact", "match the exact pattern even for ambiguous letters")
	nocomplement := opt.Switch(0, "no-complement", "do not match the complement strand")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	queries := []gts.Sequence{}
	queryBytes := []byte(*queryPath)

	h.Reset()
	switch queryBytes[0] {
	case '@':
		h.Write(queryBytes)
		query := gts.New(nil, nil, queryBytes[1:])
		queries = append(queries, query)

	default:
		queryFile, err := os.Open(*queryPath)
		if err != nil {
			return ctx.Raise(err)
		}

		r := attach(h, queryFile)
		scanner := seqio.NewAutoScanner(r)
		for scanner.Scan() {
			queries = append(queries, scanner.Value())
		}
		if len(queries) == 0 {
			ctx.Raise(fmt.Errorf("query sequence file %q does not contain a sequence", *queryPath))
		}
	}
	querySum := h.Sum(nil)

	d, err := newIODelegate(*seqinPath, *seqoutPath)
	if err != nil {
		return ctx.Raise(err)
	}
	defer d.Close()

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	order := make(map[string]int)
	props := gts.Props{}
	for _, s := range *propstrs {
		name, value := s, ""
		if i := strings.IndexByte(s, '='); i >= 0 {
			name, value = s[:i], s[i+1:]
		}
		props.Add(name, value)
		order[name] = len(order)
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"query", encodeToString(querySum)},
			{"filetype", filetype},
			{"featureKey", *featureKey},
			{"propstrs", *propstrs},
			{"exact", *exact},
			{"nocomplement", *nocomplement},
		})

		ok, err := d.TryCache(h, data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	match := gts.Match
	if *exact {
		match = gts.Search
	}

	w := bufio.NewWriter(d)

	scanner := seqio.NewAutoScanner(d)
	for scanner.Scan() {
		seq := scanner.Value()
		cmp := gts.Reverse(gts.Complement(gts.New(nil, nil, seq.Bytes())))
		ff := seq.Features()
		for _, query := range queries {
			fwd := match(seq, query)
			for _, segment := range fwd {
				head, tail := gts.Unpack(segment)
				f := gts.NewFeature(*featureKey, gts.Range(head, tail), props)
				ff = ff.Insert(f)
			}
			if !*nocomplement {
				bwd := match(cmp, query)
				for _, segment := range bwd {
					head, tail := gts.Unpack(segment)
					loc := gts.Range(head, tail)
					loc = loc.Reverse(gts.Len(seq)).(gts.Ranged)
					f := gts.NewFeature(*featureKey, loc.Complement(), props)
					ff = ff.Insert(f)
				}
			}
		}
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
