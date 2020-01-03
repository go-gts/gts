package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	flags "gopkg.in/ktnyt/flags.v1"
	gts "gopkg.in/ktnyt/gts.v0"
	pars "gopkg.in/ktnyt/pars.v2"
)

func init() {
	prog := flags.NewProgram()
	prog.Add("clear", "remove all features (excluding sources)", FeatureClear)
	prog.Add("select", "select features by feature keys", FeatureFilter)
	prog.Add("merge", "merge features from other file(s)", FeatureMerge)
	prog.Add("extract", "extract qualifier value(s) from the input record", FeatureExtract)
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
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("expected at least one record entry")
	}

	for {
		in := scanner.Record()
		ff := in.Filter(gts.Key("source"))
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

func FeatureFilter(ctx *flags.Context) error {
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
	ss := make([]gts.FeatureFilter, len(keys))
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
		ff := in.Filter(sel)
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

	defer infile.Close()
	defer outfile.Close()

	scanner := gts.NewRecordScanner(bufio.NewReader(infile))
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("expected a record entry")
	}

	in := scanner.Record()
	ff := gts.FeatureList(in.Filter())

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

func FeatureExtract(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output text file")
	mainName := pos.String("name", "qualifier name to select")

	extraNames := opt.StringSlice(0, "and", nil, "additional qualifier name(s) to select")
	delim := opt.String('d', "delimiter", "\t", "string to insert between output qualifiers")
	sep := opt.String('s', "separator", ",", "string to insert between values with same qualifier keys")
	featureKey := opt.Switch('f', "feature-key", "extract the feature key")
	location := opt.Switch('l', "location", "extract the location")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	scanner := gts.NewRecordScanner(bufio.NewReader(infile))
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("expected a record entry")
	}

	header := []string{*mainName}
	if *featureKey {
		header = append(header, "feature")
	}
	if *location {
		header = append(header, "location")
	}
	header = append(header, (*extraNames)...)
	fmt.Fprintf(outfile, "%s\n", strings.Join(header, *delim))

	in := scanner.Record()
	for _, f := range in.Filter() {
		primary := f.Qualifiers.Get(*mainName)
		if len(primary) > 0 {
			values := []string{strings.Join(primary, *sep)}
			if *featureKey {
				values = append(values, f.Key)
			}
			if *location {
				values = append(values, f.Location.String())
			}

			ok := true
			for _, name := range *extraNames {
				value := f.Qualifiers.Get(name)
				if len(value) == 0 {
					ok = false
				}
				values = append(values, strings.Join(value, *sep))
			}
			if ok {
				fmt.Fprintf(outfile, "%s\n", strings.Join(values, *delim))
			}
		}
	}

	return nil
}
