package main

import (
	"fmt"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/gt1/flags"
)

func init() {
	register("feature", featureFunc)
}

func includes(set []string, q string) bool {
	for i := range set {
		if q == set[i] {
			return true
		}
	}
	return false
}

func featureFunc(values flags.Values, parser *flags.Parser, args []string) error {
	mergeFiles := parser.Strings('m', "merge", nil, "merge the features from the given feature file(s)")
	selectKeys := parser.Strings('s', "select", nil, "select features with the given feature key(s)")
	invert := parser.Switch('v', "invert-match", "select features that do not match the given criteria")

	infile := parser.Optional("infile")
	outfile := parser.Optional("outfile")

	args, err := parser.Parse(args)
	if err != nil {
		return err
	}

	r, w := getReaderAndWriter(*infile, *outfile)

	record, err := gt1.ReadRecord(r)
	if err != nil {
		return err
	}

	features := record.Features()

	for _, filename := range *mergeFiles {
		f := Open(filename)
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
	fmt.Fprintf(w, gt1.FormatGenBank(out))

	return nil
}
