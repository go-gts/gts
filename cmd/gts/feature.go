package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	gts "gopkg.in/gts.v0"
	flags "gopkg.in/flags.v1"
	pars "gopkg.in/pars.v2"
)

func init() {
	flags.Add("clear", "remove all features (excluding sources)", FeatureClear)
	flags.Add("select", "select features by feature keys", FeatureFilter)
	flags.Add("merge", "merge features from other file(s)", FeatureMerge)
	flags.Add("extract", "extract qualifier value(s) from the input record", FeatureExtract)
	flags.Add("view", "view the input record as the specified format", FeatureView)
}

func FeatureClear(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")

	format := opt.String(0, "format", "default", "output file format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	filetype := gts.Detect(outfile.Name())
	if filetype == gts.UnknownFile {
		filetype = gts.ToFileType(*format)
	}

	scanner := gts.NewRecordFileScanner(infile)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("expected at least one record entry")
	}

	for {
		in := scanner.Record()
		ff := in.Filter(gts.Key("source"))
		out := gts.NewRecord(in.Metadata(), ff, in.Bytes())
		if _, err := gts.NewRecordFormatter(out, filetype).WriteTo(outfile); err != nil {
			return err
		}
		if !scanner.Scan() {
			break
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
	format := opt.String(0, "format", "default", "output file format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	filetype := gts.Detect(outfile.Name())
	if filetype == gts.UnknownFile {
		filetype = gts.ToFileType(*format)
	}

	keys := append([]string{*mainKey}, (*extraKeys)...)
	ss := make([]gts.FeatureFilter, len(keys))
	for i, key := range keys {
		ss[i] = gts.Key(key)
	}
	sel := gts.Or(ss...)
	if *invert {
		sel = gts.Not(sel)
	}

	scanner := gts.NewRecordFileScanner(infile)

	if !scanner.Scan() {
		return errors.New("expected at least one record entry")
	}

	for {
		in := scanner.Record()
		ff := in.Filter(sel)
		out := gts.NewRecord(in.Metadata(), ff, in.Bytes())
		if _, err := gts.NewRecordFormatter(out, filetype).WriteTo(outfile); err != nil {
			return err
		}
		if !scanner.Scan() {
			break
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
	format := opt.String(0, "format", "default", "output file format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	filetype := gts.Detect(outfile.Name())
	if filetype == gts.UnknownFile {
		filetype = gts.ToFileType(*format)
	}

	scanner := gts.NewRecordFileScanner(infile)
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

	out := gts.NewRecord(in.Metadata(), ff, in.Bytes())
	if _, err := gts.NewRecordFormatter(out, filetype).WriteTo(outfile); err != nil {
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
	format := opt.String(0, "format", "default", "output file format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	filetype := gts.Detect(outfile.Name())
	if filetype == gts.UnknownFile {
		filetype = gts.ToFileType(*format)
	}

	scanner := gts.NewRecordFileScanner(infile)
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

func FeatureView(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")

	format := opt.String(0, "format", "genbank", "output file format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	filetype := gts.Detect(outfile.Name())
	if filetype == gts.UnknownFile {
		filetype = gts.ToFileType(*format)
	}

	scanner := gts.NewRecordFileScanner(infile)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("expected at least one record entry")
	}

	for {
		rec := scanner.Record()
		if _, err := gts.NewRecordFormatter(rec, filetype).WriteTo(outfile); err != nil {
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
