package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/gt1/flags"
	"github.com/ktnyt/gt1/seqio"
)

func init() {
	register("feature", "feature manipulation commands", featureFunc)
}

func ReadOneRecord(f *os.File) (*gt1.Record, error) {
	scanner := seqio.NewScanner(f)
	if !scanner.Scan() {
		return nil, errors.New("input file cannot be interpreted as a sequence")
	}
	if record, ok := scanner.Seq().(*gt1.Record); ok {
		return record, nil
	}
	return nil, errors.New("input file is not a record")
}

func featureSelectFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input record file")
	outfile := command.Outfile("output record file")
	invert := command.Switch('v', "invert-match", "select features that do not match the given criteria")
	extraKeys := command.Strings(0, "and", "additional feature key(s) to select")
	mainKey := command.Mandatory("key", "primary feature key to select")

	return command.Run(args, func() error {
		record, err := ReadOneRecord(infile)
		if err != nil {
			return err
		}

		selectKeys := *extraKeys
		if len(*mainKey) > 0 {
			selectKeys = append(selectKeys, *mainKey)
		}

		filter := gt1.FeatureKeyFilter(selectKeys)
		if *invert {
			filter = gt1.FeatureFilterInvert(filter)
		}

		features := gt1.FilterFeatures(record.Features(), filter)

		out := gt1.NewRecord(record.Metadata(), features, record)
		fmt.Fprintf(outfile, seqio.FormatGenBank(out))

		return nil
	})
}

func featureMergeFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input record file")
	outfile := command.Outfile("output record file")
	extraFiles := command.Strings(0, "and", "additional feature file(s) to merge")
	mainFile := command.Mandatory("feature", "primary feature file to merge")

	return command.Run(args, func() error {
		record, err := ReadOneRecord(infile)
		if err != nil {
			return err
		}

		features := record.Features()

		for _, filename := range append(*extraFiles, *mainFile) {
			f, err := os.Open(filename)
			if err != nil {
				return err
			}
			tmp, err := gt1.ReadFeatures(f)
			if err != nil {
				return err
			}
			for _, feature := range tmp.Iter() {
				features.Add(feature)
			}
		}

		out := gt1.NewRecord(record.Metadata(), features, record)
		fmt.Fprintf(outfile, seqio.FormatGenBank(out))

		return nil
	})
}

func featureClearFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input record file")
	outfile := command.Outfile("output record file")
	return command.Run(args, func() error {
		record, err := ReadOneRecord(infile)
		if err != nil {
			return err
		}

		features := gt1.ClearFeatures(record.Features())

		out := gt1.NewRecord(record.Metadata(), features, record)
		fmt.Fprintf(outfile, seqio.FormatGenBank(out))

		return nil
	})
}

func qualifierJoin(vs [][]string, delim, sep string) string {
	qs := make([]string, len(vs))
	for i, v := range vs {
		qs[i] = strings.Join(v, sep)
	}
	return strings.Join(qs, delim)
}

func extractFeatureQualifiers(feature *gt1.Feature, keys []string) [][]string {
	values := make([][]string, len(keys))
	for i, key := range keys {
		value := feature.Qualifiers().All(key)
		if len(value) == 0 {
			return nil
		}
		values[i] = value
	}
	return values
}

func featureExtractFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input record file")
	outfile := command.Outfile("output text file")
	delim := command.String('d', "delimiter", "\t", "string to insert between output qualifiers")
	sep := command.String('s', "separator", ",", "string to insert between values with same qualifier keys")
	featureKey := command.Switch('f', "feature-key", "extract the feature key")
	location := command.Switch('l', "location", "extract the location")
	extraKeys := command.Strings(0, "and", "additional qualifier key(s) to extract")
	mainKey := command.Mandatory("qualifier", "primary qualifier key to extract")

	return command.Run(args, func() error {
		record, err := ReadOneRecord(infile)
		if err != nil {
			return err
		}

		features := record.Features()

		header := []string{*mainKey}
		if *featureKey {
			header = append(header, "feature")
		}
		if *location {
			header = append(header, "location")
		}
		for _, key := range *extraKeys {
			header = append(header, key)
		}

		fmt.Fprintf(outfile, "%s\n", strings.Join(header, *delim))

		for _, feature := range features.Iter() {
			primary := feature.Qualifiers().All(*mainKey)

			if len(primary) > 0 {
				values := [][]string{primary}
				if *featureKey {
					values = append(values, []string{feature.Key()})
				}
				if *location {
					values = append(values, []string{feature.Location().Format()})
				}

				extras := extractFeatureQualifiers(feature, *extraKeys)
				if extras != nil {
					values = append(values, extras...)
					fmt.Fprintf(outfile, "%s\n", qualifierJoin(values, *delim, *sep))
				}
			}
		}

		return nil
	})
}

func featureFunc(command *flags.Command, args []string) error {
	command.Command("select", "select features by feature key", featureSelectFunc)
	command.Command("merge", "merge features from a file", featureMergeFunc)
	command.Command("clear", "remove all features (excluding sources)", featureClearFunc)
	command.Command("extract", "extract qualifier value(s) from the input record", featureExtractFunc)

	return command.Run(args)
}
