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
	pos, opt := flags.Flags()

	queryString := pos.String("query", "query sequence file (will be interpreted literally if preceded with @)")

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	featureKey := opt.String('k', "key", "misc_feature", "key for the reported oligomer region features")
	exact := opt.Switch('e', "exact", "match the exact pattern even for ambiguous letters")
	nocomplement := opt.Switch(0, "no-complement", "do not match the complement strand")
	qfstrs := opt.StringSlice('q', "qualifier", nil, "qualifier key-value pairs (syntax: key=value))")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	match := gts.Match
	if *exact {
		match = gts.Search
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

	queries := []gts.Sequence{}
	queryBytes := []byte(*queryString)

	switch queryBytes[0] {
	case '@':
		query := gts.New(nil, nil, queryBytes[1:])
		queries = append(queries, query)
	default:
		queryFile, err := os.Open(*queryString)
		if err != nil {
			return ctx.Raise(err)
		}
		scanner := seqio.NewAutoScanner(queryFile)
		for scanner.Scan() {
			queries = append(queries, scanner.Value())
		}
	}

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

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

	w := bufio.NewWriter(seqoutFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		cmp := gts.Reverse(gts.Complement(gts.New(nil, nil, seq.Bytes())))
		ff := seq.Features()
		for _, query := range queries {
			fwd := match(seq, query)
			for _, segment := range append(fwd) {
				head, tail := gts.Unpack(segment)
				ff = ff.Insert(gts.Feature{
					Key:        *featureKey,
					Location:   gts.Range(head, tail),
					Qualifiers: qfs,
					Order:      order,
				})
			}
			if !*nocomplement {
				bwd := match(cmp, query)
				for _, segment := range append(bwd) {
					head, tail := gts.Unpack(segment)
					loc := gts.Range(head, tail)
					loc = loc.Reverse(gts.Len(seq)).(gts.Ranged)
					ff = ff.Insert(gts.Feature{
						Key:        *featureKey,
						Location:   loc.Complement(),
						Qualifiers: qfs,
					})
				}
			}
		}
		seq = gts.WithFeatures(seq, ff)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(w); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}
