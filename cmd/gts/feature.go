package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/flags"
	"github.com/go-gts/gts/seqio"
	"github.com/go-pars/pars"
)

func init() {
	flags.Register("clear", "remove all features from the sequence (excluding source features)", featureClear)
	flags.Register("select", "select features using the given feature selector(s)", featureSelect)
	flags.Register("annotate", "merge features from a feature list file into a sequence", featureAnnotate)
	flags.Register("query", "query information from the given sequence", featureQuery)
	flags.Register("extract", "extract the sequences referenced by the features", featureExtract)
}

func featureClear(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
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
		ff := seq.Features().Filter(gts.Key("source"))
		seq = gts.WithFeatures(seq, ff)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(seqoutFile); err != nil {
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

func featureSelect(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	selector := pos.String("selector", "feature selector (syntax: [feature_key][/[qualifier1][=regexp1]][/[qualifier2][=regexp2]]...)")

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	invert := opt.Switch('v', "invert-match", "select features that do not match the given criteria")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	filter, err := gts.Selector(*selector)
	if err != nil {
		return ctx.Raise(fmt.Errorf("invalid selector syntax: %v", err))
	}
	if *invert {
		filter = gts.Not(filter)
	}
	filter = gts.Or(gts.Key("source"), filter)

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
		ff := seq.Features().Filter(filter)
		seq = gts.WithFeatures(seq, ff)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(seqoutFile); err != nil {
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

func featureAnnotate(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	featinPath := pos.String("feature_table", "feature table file containing features to merge")

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
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

	featinFile, err := os.Open(*featinPath)
	if err != nil {
		return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *featinPath, err))
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

	state := pars.NewState(featinFile)
	result, err := gts.FeatureTableParser("").Parse(state)
	featin := result.Value.(gts.FeatureTable)

	w := bufio.NewWriter(seqoutFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		ff := seq.Features()
		for _, f := range featin {
			ff = ff.Insert(f)
		}
		seq = gts.WithFeatures(seq, ff)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(seqoutFile); err != nil {
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

func featureQuery(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

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

	sep := []rune(*sepstr)
	if len(sep) > 1 {
		return ctx.Raise(fmt.Errorf("separator must be a single character: got %q", *sepstr))
	}
	comma := sep[0]

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

	outFile := os.Stdout
	if *outPath != "-" {
		f, err := os.Create(*outPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *outPath, err))
		}
		outFile = f
		defer outFile.Close()
	}

	w := bufio.NewWriter(outFile)

	var common []string = nil

	ids := make([]string, 0)
	fff := [][]gts.Feature{}

	scanner := seqio.NewAutoScanner(seqinFile)
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
			for i := 0; i < len(common); i++ {
				if _, ok := f.Qualifiers[common[i]]; !ok {
					common = append(common[:i], common[i+1:]...)
				}
			}
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

func featureExtract(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	modstr := opt.String('m', "--range", "^..$", "location range modifier")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	mod, err := gts.AsModifier(*modstr)
	if err != nil {
		return ctx.Raise(fmt.Errorf("bad range modifier: %v", err))
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
		ff := seq.Features().Filter(gts.Not(gts.Key("source")))
		for _, f := range ff {
			region := f.Location.Region()
			region = region.Resize(mod)
			out := region.Locate(seq)
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
