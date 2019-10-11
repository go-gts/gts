package main

import (
	"fmt"
	"os"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/gt1/flags"
)

func init() {
	desc := "manipulate feature table"
	register("feature", desc, featureFunc)
}

func featureSelectFunc(command *flags.Command, args []string) error {
	invert := command.Switch('v', "invert-match", "select features that do not match the given criteria")
	extraKeys := command.Strings('a', "and", "additional feature key(s) to select")
	mainKey := command.Mandatory("key", "feature key to select")

	infile := command.Infile("input record file")
	outfile := command.Outfile("output record file")

	return command.Run(args, func() error {
		record, err := gt1.ReadRecord(infile)
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

		out := gt1.NewRecord(record.Fields(), features, record)
		fmt.Fprintf(outfile, gt1.FormatGenBank(out))

		return nil
	})
}

func featureMergeFunc(command *flags.Command, args []string) error {
	extraFiles := command.Strings('a', "and", "additional feature file(s) to merge")
	mainFile := command.Mandatory("feature", "feature file to merge")

	infile := command.Infile("input record file")
	outfile := command.Outfile("output record file")

	return command.Run(args, func() error {
		record, err := gt1.ReadRecord(infile)
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
			features = append(features, tmp...)
		}

		out := gt1.NewRecord(record.Fields(), features, record)
		fmt.Fprintf(outfile, gt1.FormatGenBank(out))

		return nil
	})
}

func featureClearFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input record file")
	outfile := command.Outfile("output record file")

	return command.Run(args, func() error {
		record, err := gt1.ReadRecord(infile)
		if err != nil {
			return err
		}

		features := gt1.ClearFeatures(record.Features())

		out := gt1.NewRecord(record.Fields(), features, record)
		fmt.Fprintf(outfile, gt1.FormatGenBank(out))

		return nil
	})
}

func featureFunc(command *flags.Command, args []string) error {
	command.Command("select", "select features by feature key", featureSelectFunc)
	command.Command("merge", "merge features from a file", featureMergeFunc)
	command.Command("clear", "remove all features (excluding sources)", featureClearFunc)

	return command.Run(args)
}
