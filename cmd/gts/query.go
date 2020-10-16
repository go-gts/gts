package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("query", "query information from the given sequence", queryFunc)
}

func formatCSV(record []string, comma rune) (string, error) {
	b := strings.Builder{}
	w := csv.NewWriter(&b)
	w.Comma = comma
	if err := w.Write(record); err != nil {
		return "", err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	s := strings.TrimSpace(b.String())
	s = strings.ReplaceAll(s, "\n", " ")
	return s, nil
}

func queryFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	outPath := opt.String('o', "output", "-", "output table file (specifying `-` will force standard output)")
	names := opt.StringSlice('n', "name", nil, "qualifier name(s) to select")
	delim := opt.String('d', "delimiter", "\t", "string to insert between columns")
	sepstr := opt.String('t', "separator", ",", "string to insert between qualifier values")
	noheader := opt.Switch(0, "no-header", "do not print the header line")
	source := opt.Switch(0, "source", "include the source feature(s)")
	nokey := opt.Switch(0, "no-key", "do not report the feature key")
	noloc := opt.Switch(0, "no-location", "do not report the feature location")
	empty := opt.Switch(0, "empty", "allow missing qualifiers to be reported")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	d, err := newIODelegate(h, *seqinPath, *outPath)
	if err != nil {
		return ctx.Raise(err)
	}
	defer d.Close()

	sort.Strings(*names)

	sep := []rune(*sepstr)
	if len(sep) > 1 {
		return ctx.Raise(fmt.Errorf("separator must be a single character: got %q", *sepstr))
	}
	comma := sep[0]

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"names", *names},
			{"delim", *delim},
			{"comma", comma},
			{"noheader", *noheader},
			{"source", *source},
			{"nokey", *nokey},
			{"noloc", *noloc},
			{"empty", *empty},
		})

		ok, err := d.Cache(data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	var common []string = nil

	ids := make([]string, 0)
	fff := [][]gts.Feature{}

	w := bufio.NewWriter(d)

	scanner := seqio.NewAutoScanner(d)
	for scanner.Scan() {
		seq := scanner.Value()
		switch info := seq.Info().(type) {
		case interface{ ID() string }:
			ids = append(ids, info.ID())
		default:
			ids = append(ids, "")
		}
		ff := seq.Features()
		for _, f := range ff {
			if !*source && f.Key == "source" {
				continue
			}
			if common == nil {
				for key := range f.Qualifiers {
					common = append(common, key)
				}
				sort.Strings(common)
			}
			indices := make([]int, 0, len(common))
			for i := 0; i < len(common); i++ {
				if _, ok := f.Qualifiers[common[i]]; ok {
					indices = append(indices, i)
				}
			}
			remain := make([]string, len(indices))
			for i, index := range indices {
				remain[i] = common[index]
			}
			common = remain
		}
		fff = append(fff, ff)
	}

	if len(*names) > 0 {
		common = *names
	}

	if !*noheader {
		fields := []string{}
		fields = append(fields, "seq")
		if !*nokey {
			fields = append(fields, "feature")
		}
		if !*noloc {
			fields = append(fields, "location")
		}
		fields = append(fields, common...)
		header := fmt.Sprintf("%s\n", strings.Join(fields, *delim))
		_, err := io.WriteString(w, header)
		if err != nil {
			return ctx.Raise(err)
		}
	}

	n := len(fmt.Sprintf("%d", len(fff)))
	format := fmt.Sprintf("%%0%dd", n)

	for i, ff := range fff {
		id := ids[i]
		if id == "" {
			id = fmt.Sprintf(format, i)
		}

		for _, f := range ff {
			cc := []string{id}

			if !*nokey {
				cc = append(cc, f.Key)
			}
			if !*noloc {
				cc = append(cc, f.Location.String())
			}

			ok := (*source || f.Key != "source")

			for _, name := range common {
				vv := f.Qualifiers.Get(name)
				if len(vv) == 0 && !*empty {
					ok = false
				}
				s, err := formatCSV(vv, comma)
				if err != nil {
					return ctx.Raise(err)
				}
				cc = append(cc, s)
			}

			if ok {
				line := fmt.Sprintf("%s\n", strings.Join(cc, *delim))
				_, err := io.WriteString(w, line)
				if err != nil {
					return ctx.Raise(err)
				}
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
