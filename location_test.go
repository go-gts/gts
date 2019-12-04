package gt1_test

import (
	"fmt"
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

func ints(is ...int) []int { return is }

func forceLocation(s string) gt1.Location {
	loc, err := gt1.AsLocation(s)
	if err != nil {
		panic(err)
	}
	return loc
}

func TestLocationLess(t *testing.T) {
	values := []string{
		"0",
		"0..42",
		"42",
		"42..723",
		"723",
	}
	cases := make([]assert.F, len(values))

	for i := range values {
		tmp := make([]assert.F, len(values))
		for j := range values {
			a, b := forceLocation(values[i]), forceLocation(values[j])
			tmp[j] = assert.All(
				assert.Equal(gt1.LocationLess(a, b), i < j),
				assert.Equal(gt1.LocationLess(b, a), j < i),
			)
		}
		cases[i] = assert.All(tmp...)
	}

	assert.Apply(t, cases...)
}

func TestLocationsLocate(t *testing.T) {
	seq := gt1.Seq("atgcatgc")

	values := []struct {
		Loc  gt1.Location
		ESeq gt1.Sequence
		EIdx []int
	}{
		{gt1.NewPointLocation(0), gt1.Seq("a"), ints(0)},
		{gt1.NewPointLocation(1), gt1.Seq("t"), ints(1)},
		{gt1.NewPointLocation(2), gt1.Seq("g"), ints(2)},
		{gt1.NewPointLocation(3), gt1.Seq("c"), ints(3)},

		{gt1.NewRangeLocation(0, 4), gt1.Seq("atgc"), ints(0, 1, 2, 3)},
		{gt1.NewRangeLocation(2, 6), gt1.Seq("gcat"), ints(2, 3, 4, 5)},
		{gt1.NewRangeLocation(4, 8), gt1.Seq("atgc"), ints(4, 5, 6, 7)},

		{gt1.NewAmbiguousLocation(0, 4), gt1.Seq("atgc"), ints(0, 1, 2, 3)},
		{gt1.NewAmbiguousLocation(2, 6), gt1.Seq("gcat"), ints(2, 3, 4, 5)},
		{gt1.NewAmbiguousLocation(4, 8), gt1.Seq("atgc"), ints(4, 5, 6, 7)},

		{gt1.NewBetweenLocation(0, 4), gt1.Seq("atgc"), ints(0, 1, 2, 3)},
		{gt1.NewBetweenLocation(2, 6), gt1.Seq("gcat"), ints(2, 3, 4, 5)},
		{gt1.NewBetweenLocation(4, 8), gt1.Seq("atgc"), ints(4, 5, 6, 7)},

		{
			gt1.NewComplementLocation(gt1.NewRangeLocation(0, 4)),
			gt1.Complement(gt1.Seq("atgc")), ints(0, 1, 2, 3),
		},
		{
			gt1.NewComplementLocation(gt1.NewRangeLocation(2, 6)),
			gt1.Complement(gt1.Seq("gcat")), ints(2, 3, 4, 5),
		},
		{
			gt1.NewComplementLocation(gt1.NewRangeLocation(4, 8)),
			gt1.Complement(gt1.Seq("atgc")), ints(4, 5, 6, 7),
		},

		{
			gt1.NewJoinLocation([]gt1.Location{
				gt1.NewRangeLocation(0, 2),
				gt1.NewRangeLocation(3, 5),
				gt1.NewRangeLocation(6, 8),
			}),
			gt1.Seq("atcagc"),
			ints(0, 1, 3, 4, 6, 7),
		},
		{
			gt1.NewJoinLocation([]gt1.Location{
				gt1.NewRangeLocation(0, 3),
				gt1.NewRangeLocation(5, 8),
			}),
			gt1.Seq("atgtgc"),
			ints(0, 1, 2, 5, 6, 7),
		},

		{
			gt1.NewOrderLocation([]gt1.Location{
				gt1.NewRangeLocation(0, 2),
				gt1.NewRangeLocation(3, 5),
				gt1.NewRangeLocation(6, 8),
			}),
			gt1.Seq("atcagc"),
			ints(0, 1, 3, 4, 6, 7),
		},
		{
			gt1.NewOrderLocation([]gt1.Location{
				gt1.NewRangeLocation(0, 3),
				gt1.NewRangeLocation(5, 8),
			}),
			gt1.Seq("atgtgc"),
			ints(0, 1, 2, 5, 6, 7),
		},
	}

	cases := make([]assert.F, len(values))
	for i, value := range values {
		loc, eseq, eidx := value.Loc, value.ESeq, value.EIdx
		idx := make([]int, loc.Len())
		for i := range idx {
			idx[i] = loc.Map(i)
		}

		cases[i] = assert.All(
			assert.Equal(loc.Len(), len(eseq.Bytes())),
			assert.Equal(loc.Locate(seq), eseq),
			assert.Equal(idx, eidx),
		)
	}

	assert.Apply(t, cases...)
}

func testLocationShift(s0, s1, s2 string, valid bool) assert.F {
	name := fmt.Sprintf("Location(%s)", s0)

	zero := func() assert.F {
		loc, exp := forceLocation(s0), forceLocation(s0)
		return assert.All(
			assert.True(loc.Shift(1, 0)),
			assert.Equal(loc.String(), exp.String()),
		)
	}

	up := func() assert.F {
		loc, exp := forceLocation(s0), forceLocation(s1)
		return assert.All(
			assert.True(loc.Shift(1, 1)),
			assert.Equal(loc.String(), exp.String()),
		)
	}

	down := func() assert.F {
		loc, exp := forceLocation(s0), forceLocation(s2)
		return assert.All(
			assert.Equal(loc.Shift(1, -1), valid),
			assert.Equal(loc.String(), exp.String()),
		)
	}

	return assert.C(name,
		assert.C("shift zero", zero()),
		assert.C("shift up", up()),
		assert.C("shift down", down()),
	)
}

func TestLocationsShift(t *testing.T) {
	values := []struct {
		s0, s1, s2 string
		valid      bool
	}{
		{"1", "1", "1", true},
		{"2", "3", "2", false},
		{"3", "4", "2", true},

		{"1..2", "1..3", "1..2", false},
		{"1..3", "1..4", "1..2", true},
		{"2..3", "3..4", "2..3", false},
		{"2..4", "3..5", "2..3", true},
		{"3..4", "4..5", "2..3", true},

		{"1.2", "1.3", "1.2", false},
		{"1.3", "1.4", "1.2", true},
		{"2.3", "3.4", "2.3", false},
		{"2.4", "3.5", "2.3", true},
		{"3.4", "4.5", "2.3", true},

		{"1^2", "1^3", "1^2", false},
		{"1^3", "1^4", "1^2", true},
		{"2^3", "3^4", "2^3", false},
		{"2^4", "3^5", "2^3", true},
		{"3^4", "4^5", "2^3", true},

		{"complement(1..2)", "complement(1..3)", "complement(1..2)", false},
		{"complement(1..3)", "complement(1..4)", "complement(1..2)", true},
		{"complement(2..3)", "complement(3..4)", "complement(2..3)", false},
		{"complement(2..4)", "complement(3..5)", "complement(2..3)", true},
		{"complement(3..4)", "complement(4..5)", "complement(2..3)", true},

		{"join(1..2,2..3)", "join(1..3,3..4)", "join(1..2,2..3)", false},
		{"join(1..3,3..4)", "join(1..4,4..5)", "join(1..2,2..3)", true},

		{"order(1..2,2..3)", "order(1..3,3..4)", "order(1..2,2..3)", false},
		{"order(1..3,3..4)", "order(1..4,4..5)", "order(1..2,2..3)", true},
	}

	cases := make([]assert.F, len(values))
	for i, value := range values {
		s0, s1, s2, valid := value.s0, value.s1, value.s2, value.valid
		cases[i] = testLocationShift(s0, s1, s2, valid)
	}

	assert.Apply(t, cases...)
}

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
				assert.Equal(r.Value.(gt1.Location).String(), e),
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

func TestLocationIO(t *testing.T) {
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
