package gts

import (
	"encoding/json"
	"fmt"
	"io"

	humanize "github.com/dustin/go-humanize"
	pars "gopkg.in/ktnyt/pars.v2"
	yaml "gopkg.in/yaml.v2"
)

// FeatureSelector represents a filter for selecting features.
type FeatureSelector func(f Feature) bool

// Any returns true for any feature matched.
func Any(f Feature) bool { return true }

// Key returns true if the key of a feature matches the given key.
func Key(key string) FeatureSelector {
	return func(f Feature) bool { return f.Key == key }
}

// And returns true if all selectors return true.
func And(ss ...FeatureSelector) FeatureSelector {
	return func(f Feature) bool {
		for _, s := range ss {
			if !s(f) {
				return false
			}
		}
		return true
	}
}

// Or returns true if any of the selectors return true.
func Or(ss ...FeatureSelector) FeatureSelector {
	return func(f Feature) bool {
		for _, s := range ss {
			if s(f) {
				return true
			}
		}
		return false
	}
}

// Not returns true if the given selector returns false.
func Not(s FeatureSelector) FeatureSelector {
	return func(f Feature) bool { return !s(f) }
}

// FeatureTable represents a feature table.
type FeatureTable interface {
	Select(ss ...FeatureSelector) []Feature
	Add(f Feature)
}

type featureIO struct {
	Key        *string
	Location   *string
	Qualifiers [][]string
}

func featureDecoderParser(f newDecoder) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		d := f(state)
		pfios := new([]featureIO)
		state.Push()
		if err := d.Decode(pfios); err != nil {
			state.Pop()
			return err
		}
		state.Drop()

		ff := make([]Feature, len(*pfios))
		for i, fio := range *pfios {
			ord := humanize.Ordinal(i + 1)
			if fio.Key == nil {
				return fmt.Errorf("%s feature is missing a key", ord)
			}
			key := *fio.Key
			if fio.Location == nil {
				return fmt.Errorf("%s feature is missing a location", ord)
			}
			s := *fio.Location
			loc, err := AsLocation(s)
			if err != nil {
				return fmt.Errorf("%s feature location string %q cannot be parsed", ord, s)
			}
			qfs := Values{}
			for _, item := range fio.Qualifiers {
				qfs.Add(item[0], item[1])
			}
			ff[i] = NewFeature(key, loc, qfs)
		}
		result.SetValue(ff)
		return nil
	}
}

func yamlDecoder(r io.Reader) decoder { return yaml.NewDecoder(r) }
func jsonDecoder(r io.Reader) decoder { return json.NewDecoder(r) }

// FeatureTableParser attempts to parse a table of features.
var FeatureTableParser = pars.Any(
	RecordParser.Map(func(result *pars.Result) error {
		rec := result.Value.(Record)
		ft := rec.Select()
		result.SetValue(ft)
		return nil
	}),
	featureDecoderParser(yamlDecoder),
	featureDecoderParser(jsonDecoder),
)

// ReadFeatureTable attempts to read and parse a table of features.
func ReadFeatureTable(r io.Reader) (FeatureList, error) {
	result, err := FeatureTableParser.Parse(pars.NewState(r))
	if err != nil {
		return nil, err
	}
	return result.Value.(FeatureList), nil
}
