package main

import (
	"bufio"
	"bytes"
	"errors"
	"os"

	"github.com/ktnyt/gts"
	flags "gopkg.in/ktnyt/flags.v1"
	pars "gopkg.in/ktnyt/pars.v2"
)

func init() {
	prog := flags.NewProgram()
	prog.Add("clear", "remove all features (excluding sources)", FeatureClear)
	prog.Add("select", "select features by feature keys", FeatureSelect)
	prog.Add("merge", "merge features from other file(s)", FeatureMerge)
	flags.Add("feature", "feature manipulation commands", prog.Compile())
}

func FeatureClear(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	scanner := gts.NewRecordScanner(bufio.NewReader(infile))

	if !scanner.Scan() {
		return errors.New("expected at least one record entry")
	}

	for {
		in := scanner.Record()
		ff := in.Select(gts.Key("source"))
		buffer := &bytes.Buffer{}
		out := gts.NewRecord(in.Metadata(), ff, in.Bytes())
		gts.DefaultFormatter(out).WriteTo(buffer)

		if _, err := buffer.WriteTo(outfile); err != nil {
			return err
		}

		if !scanner.Scan() {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func FeatureSelect(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")
	mainKey := pos.String("key", "feature key to select")

	invert := opt.Switch('v', "invert-match", "select feature that do not match the given criteria")
	extraKeys := opt.StringSlice(0, "and", nil, "additional feature key(s) to select")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	keys := append([]string{*mainKey}, (*extraKeys)...)
	ss := make([]gts.FeatureSelector, len(keys))
	for i, key := range keys {
		ss[i] = gts.Key(key)
	}
	sel := gts.Or(ss...)
	if *invert {
		sel = gts.Not(sel)
	}

	scanner := gts.NewRecordScanner(bufio.NewReader(infile))

	if !scanner.Scan() {
		return errors.New("expected at least one record entry")
	}

	for {
		in := scanner.Record()
		ff := in.Select(sel)
		buffer := &bytes.Buffer{}
		out := gts.NewRecord(in.Metadata(), ff, in.Bytes())
		gts.DefaultFormatter(out).WriteTo(buffer)

		if _, err := buffer.WriteTo(outfile); err != nil {
			return err
		}

		if !scanner.Scan() {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func FeatureMerge(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")
	mainFile := pos.Open("feature", "primary feature file to merge")

	extraFiles := opt.OpenSlice(0, "and", nil, "additional feature file(s) to merge")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	scanner := gts.NewRecordScanner(bufio.NewReader(infile))
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("expected a record entry")
	}

	in := scanner.Record()
	ff := gts.FeatureList(in.Select(gts.Any))

	files := append([]*os.File{mainFile}, (*extraFiles)...)
	for _, f := range files {
		state := pars.NewState(f)
		result, err := gts.FeatureTableParser.Parse(state)
		if err != nil {
			return err
		}
		features := result.Value.([]gts.Feature)
		for _, feat := range features {
			ff.Add(feat)
		}
	}

	buffer := &bytes.Buffer{}
	out := gts.NewRecord(in.Metadata(), ff, in.Bytes())
	gts.DefaultFormatter(out).WriteTo(buffer)

	if _, err := buffer.WriteTo(outfile); err != nil {
		return err
	}

	return nil
}
