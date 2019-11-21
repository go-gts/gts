package gt1_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

var locationStrings = []struct {
	Name  string
	Value string
}{
	{"PointLocationParser", "42"},

	{"RangeLocationParser", "1..42"},
	{"RangeLocationParser", "<1..42"},
	{"RangeLocationParser", "1..>42"},
	{"RangeLocationParser", "<1..>42"},

	{"AmbiguousLocationParser", "1.42"},

	{"BetweenLocationParser", "1^42"},

	{"ComplementLocationParser", "complement(1..42)"},
	{"ComplementLocationParser", "complement(<1..42)"},
	{"ComplementLocationParser", "complement(1..>42)"},
	{"ComplementLocationParser", "complement(<1..>42)"},
	{"ComplementLocationParser", "complement(join(1..42,346..723))"},
	{"ComplementLocationParser", "complement(join(<1..42,346..723))"},
	{"ComplementLocationParser", "complement(join(1..>42,346..723))"},
	{"ComplementLocationParser", "complement(join(<1..>42,346..723))"},
	{"ComplementLocationParser", "complement(join(1..42,<346..723))"},
	{"ComplementLocationParser", "complement(join(1..42,346..>723))"},
	{"ComplementLocationParser", "complement(join(1..42,<346..>723))"},
	{"ComplementLocationParser", "complement(join(<1..42,<346..723))"},
	{"ComplementLocationParser", "complement(join(1..>42,346..>723))"},
	{"ComplementLocationParser", "complement(join(<1..>42,<346..>723))"},
	{"ComplementLocationParser", "complement(join(1..42,complement(346..723)))"},
	{"ComplementLocationParser", "complement(join(complement(1..42),346..723))"},
	{"ComplementLocationParser", "complement(join(complement(1..42),complement(346..723)))"},

	{"JoinLocationParser", "join(1..42,346..723)"},
	{"JoinLocationParser", "join(<1..42,346..723)"},
	{"JoinLocationParser", "join(1..>42,346..723)"},
	{"JoinLocationParser", "join(<1..>42,346..723)"},
	{"JoinLocationParser", "join(1..42,<346..723)"},
	{"JoinLocationParser", "join(1..42,346..>723)"},
	{"JoinLocationParser", "join(1..42,<346..>723)"},
	{"JoinLocationParser", "join(<1..42,<346..723)"},
	{"JoinLocationParser", "join(1..>42,346..>723)"},
	{"JoinLocationParser", "join(<1..>42,<346..>723)"},
	{"JoinLocationParser", "join(1..42,complement(346..723))"},
	{"JoinLocationParser", "join(complement(1..42),346..723)"},
	{"JoinLocationParser", "join(complement(1..42),complement(346..723))"},

	{"OrderLocationParser", "order(1..42,346..723)"},
	{"OrderLocationParser", "order(<1..42,346..723)"},
	{"OrderLocationParser", "order(1..>42,346..723)"},
	{"OrderLocationParser", "order(<1..>42,346..723)"},
	{"OrderLocationParser", "order(1..42,<346..723)"},
	{"OrderLocationParser", "order(1..42,346..>723)"},
	{"OrderLocationParser", "order(1..42,<346..>723)"},
	{"OrderLocationParser", "order(<1..42,<346..723)"},
	{"OrderLocationParser", "order(1..>42,346..>723)"},
	{"OrderLocationParser", "order(<1..>42,<346..>723)"},
	{"OrderLocationParser", "order(1..42,complement(346..723))"},
	{"OrderLocationParser", "order(complement(1..42),346..723)"},
	{"OrderLocationParser", "order(complement(1..42),complement(346..723))"},
}

var locationNameMap = map[string]pars.Parser{
	"PointLocationParser":      gt1.PointLocationParser,
	"RangeLocationParser":      gt1.RangeLocationParser,
	"AmbiguousLocationParser":  gt1.AmbiguousLocationParser,
	"BetweenLocationParser":    gt1.BetweenLocationParser,
	"OrderLocationParser":      gt1.OrderLocationParser,
	"JoinLocationParser":       gt1.JoinLocationParser,
	"ComplementLocationParser": gt1.ComplementLocationParser,
}

func testLocationStrings(name string) assert.F {
	validCases, invalidCases := []assert.F{}, []assert.F{}

	p := pars.Exact(locationNameMap[name])

	for _, pair := range locationStrings {
		e := pair.Value
		s := pars.FromString(e)

		if name == pair.Name {
			r := pars.Result{}

			validCases = append(validCases, assert.All(
				assert.NoError(p(s, &r)),
				assert.Equal(r.Value.(gt1.Location).Format(), e),
			))
		} else {
			invalidCases = append(invalidCases, assert.IsError(p(s, pars.Void)))
		}
	}

	return assert.C(name,
		assert.C("valid", validCases...),
		assert.C("invalid", invalidCases...),
	)
}

func TestLocationParsers(t *testing.T) {
	parsers := []string{
		"ComplementLocationParser",
		"PointLocationParser",
		"RangeLocationParser",
		"AmbiguousLocationParser",
		"BetweenLocationParser",
		"OrderLocationParser",
		"JoinLocationParser",
	}

	cases := make([]assert.F, len(parsers))

	for i, parser := range parsers {
		cases[i] = testLocationStrings(parser)
	}

	assert.Apply(t, cases...)
}
