package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	flags "gopkg.in/flags.v1"
	gts "gopkg.in/gts.v0"
	pars "gopkg.in/pars.v2"
)

func init() {
	flags.Add("clear", "remove all features (excluding sources)", FeatureClear)
	flags.Add("select", "select features by feature keys", FeatureSelect)
	flags.Add("merge", "merge features from other file(s)", FeatureMerge)
	flags.Add("extract", "extract qualifier value(s) from the given record", FeatureExtract)
	flags.Add("view", "view the input record as the specified format", FeatureView)
	flags.Add("seq", "retrieve the feature sequences from the given record", FeatureSeq)
}

func FeatureClear(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")

	format := opt.String('F', "format", "default", "output file format")

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

	for scanner.Scan() {
		in := scanner.Record()
		ff := in.Filter(gts.Key("source"))
		out := gts.NewRecord(in, ff)
		if _, err := gts.NewRecordWriter(out, filetype).WriteTo(outfile); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(err)
	}

	if scanner.Record() == nil {
		return ctx.Raise(errors.New("expected at least one record entry"))
	}

	return nil
}

func FeatureSelect(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")
	selector := pos.String("selector", "feature selector")

	invert := opt.Switch('v', "invert-match", "select feature that do not match the given criteria")
	format := opt.String('F', "format", "default", "output file format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	defer infile.Close()
	defer outfile.Close()

	filter := gts.ParseSelector(*selector)
	if *invert {
		filter = gts.Not(filter)
	}

	filetype := gts.Detect(outfile.Name())
	if filetype == gts.UnknownFile {
		filetype = gts.ToFileType(*format)
	}

	scanner := gts.NewRecordFileScanner(infile)

	for scanner.Scan() {
		in := scanner.Record()
		ff := in.Filter(filter)
		out := gts.NewRecord(in, ff)
		if _, err := gts.NewRecordWriter(out, filetype).WriteTo(outfile); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(err)
	}

	if scanner.Record() == nil {
		return ctx.Raise(errors.New("expected at least one record entry"))
	}

	return nil
}

func FeatureMerge(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")
	mainFile := pos.Open("feature", "primary feature file to merge")

	extraFiles := opt.OpenSlice(0, "and", nil, "additional feature file(s) to merge")
	format := opt.String('F', "format", "default", "output file format")

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
			return ctx.Raise(err)
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
			return ctx.Raise(err)
		}
		features := result.Value.([]gts.Feature)
		for _, feat := range features {
			ff.Add(feat)
		}
	}

	out := gts.NewRecord(in, ff)
	if _, err := gts.NewRecordWriter(out, filetype).WriteTo(outfile); err != nil {
		return ctx.Raise(err)
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
	sep := opt.String('t', "separator", ",", "string to insert between values with same qualifier keys")
	featureKey := opt.Switch('f', "feature-key", "extract the feature key")
	location := opt.Switch('l', "location", "extract the location")
	format := opt.String('F', "format", "default", "output file format")

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
	_, err := io.WriteString(outfile, strings.Join(header, *delim)+"\n")
	if err != nil {
		return ctx.Raise(err)
	}

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

	format := opt.String('F', "format", "genbank", "output file format")

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

	for scanner.Scan() {
		rec := scanner.Record()
		if _, err := gts.NewRecordWriter(rec, filetype).WriteTo(outfile); err != nil {
			return ctx.Raise(err)
		}
		if !scanner.Scan() {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(err)
	}

	if scanner.Record() == nil {
		return ctx.Raise(errors.New("expected at least one record entry"))
	}

	return nil
}

func FeatureSeq(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output sequence file")
	qualifier := pos.String("qualifier", "qualifier name to use for sequence description")

	and := opt.StringSlice(0, "and", nil, "additional qualifier names to use for sequence description")
	delim := opt.String('d', "delim", "|", "string to insert between description values")
	format := opt.String('F', "format", "genbank", "output file format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	filetype := gts.Detect(outfile.Name())
	if filetype == gts.UnknownFile {
		filetype = gts.ToFileType(*format)
	}

	scanner := gts.NewRecordFileScanner(infile)

	for scanner.Scan() {
		rec := scanner.Record()
		for _, f := range rec.Filter() {
			b := strings.Builder{}
			values, ok := f.Qualifiers[*qualifier]
			if ok {
				b.WriteString(values[0])
				for _, value := range values[1:] {
					b.WriteString(*delim)
					b.WriteString(value)
				}

				for _, name := range *and {
					for _, value := range f.Qualifiers[name] {
						b.WriteString(*delim)
						b.WriteString(value)
					}
				}
				seq := gts.New(b.String(), f.Bytes())
				if _, err := gts.NewSequenceWriter(seq, filetype).WriteTo(outfile); err != nil {
					return ctx.Raise(err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(err)
	}

	if scanner.Record() == nil {
		return ctx.Raise(errors.New("expected at least one record entry"))
	}

	return nil
}
