package main

import (
	"fmt"
	"os"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/gt1/flags"
)

func init() {
	register("feature", "manipulate features", featureFunc)
}

func includes(set []string, q string) bool {
	for i := range set {
		if q == set[i] {
			return true
		}
	}
	return false
}

func featureFunc(command *flags.Command, args []string) error {
	mergeFiles := command.Strings('m', "merge", "merge the features from the given feature file(s)")
	selectKeys := command.Strings('s', "select", "select features with the given feature key(s)")
	invert := command.Switch('v', "invert-match", "select features that do not match the given criteria")

	infile := command.Infile()
	outfile := command.Outfile()

	if err := command.Run(args); err != nil {
		return err
	}

	record, err := gt1.ReadRecord(infile)
	if err != nil {
		return err
	}

	features := record.Features()

	for _, filename := range *mergeFiles {
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

	filter := gt1.FeatureFilterAnd(
		gt1.FeatureKeyFilter(*selectKeys),
	)

	if *invert {
		filter = gt1.FeatureFilterInvert(filter)
	}

	filtered := gt1.FilterFeatures(features, filter)

	out := gt1.NewRecord(record.Fields(), filtered, record)
	fmt.Fprintf(outfile, gt1.FormatGenBank(out))

	return nil
}
