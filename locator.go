package gts

import (
	"errors"
	"strings"

	"github.com/go-pars/pars"
)

// Locator is a function that maps features to its regions.
type Locator func(ff FeatureTable) Regions

func allLocator(ff FeatureTable) Regions {
	rr := make(Regions, len(ff))
	for i, f := range ff {
		rr[i] = f.Location.Region()
	}
	return rr
}

func resizeLocator(locate Locator, mod Modifier) Locator {
	return func(ff FeatureTable) Regions {
		rr := locate(ff)
		for i, r := range rr {
			rr[i] = r.Resize(mod)
		}
		return rr
	}
}

func locationLocator(loc Location) Locator {
	return func(ff FeatureTable) Regions {
		return Regions{loc.Region()}
	}
}

func filterLocator(f Filter) Locator {
	return func(ff FeatureTable) Regions {
		ff = ff.Filter(f)
		rr := make(Regions, len(ff))
		for i, f := range ff {
			rr[i] = f.Location.Region()
		}
		return rr
	}
}

func tryLocation(s string) (Location, bool) {
	var parser pars.Parser
	parser = pars.Any(parseComplement(&parser), parseRange, parsePoint)
	result, err := parser.Parse(pars.FromString(s))
	if err != nil {
		return nil, false
	}
	return result.Value.(Location), true
}

// AsLocator interprets the given string as a Locator.
func AsLocator(s string) (Locator, error) {
	switch i := strings.IndexByte(s, '@'); i {
	case -1:
		loc, ok := tryLocation(s)
		if ok {
			return locationLocator(loc), nil
		}

		sel, err := Selector(s)
		if err == nil {
			return filterLocator(sel), nil
		}

		return nil, errors.New("expected a selector or locator")
	case 0:
		mod, err := AsModifier(s[1:])
		if err != nil {
			return nil, err
		}
		return resizeLocator(allLocator, mod), nil

	default:
		locate, err := AsLocator(s[:i])
		if err != nil {
			return nil, err
		}
		mod, err := AsModifier(s[i+1:])
		if err != nil {
			return nil, err
		}
		return resizeLocator(locate, mod), nil
	}
}
