package gts_test

import (
	"fmt"
	"testing"

	"gopkg.in/ktnyt/assert.v1"
	"gopkg.in/ktnyt/gts.v0"
	"gopkg.in/ktnyt/pars.v2"
)

func ints(is ...int) []int { return is }

func forceLocation(s string) gts.Location {
	loc, err := gts.AsLocation(s)
	if err != nil {
		panic(err)
	}
	return loc
}

func TestLocationLess(t *testing.T) {
	values := []string{
		"1",
		"1..42",
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
				assert.Equal(gts.LocationLess(a, b), i < j),
				assert.Equal(gts.LocationLess(b, a), j < i),
			)
		}
		cases[i] = assert.All(tmp...)
	}

	assert.Apply(t, cases...)
}

func TestLocationsLocate(t *testing.T) {
	seq := gts.Seq("atgcatgc")

	values := []struct {
		Loc  gts.Location
		ESeq gts.Sequence
		EIdx []int
	}{
		{gts.NewPointLocation(0), gts.Seq("a"), ints(0)},
		{gts.NewPointLocation(1), gts.Seq("t"), ints(1)},
		{gts.NewPointLocation(2), gts.Seq("g"), ints(2)},
		{gts.NewPointLocation(3), gts.Seq("c"), ints(3)},

		{gts.NewRangeLocation(0, 4), gts.Seq("atgc"), ints(0, 1, 2, 3)},
		{gts.NewRangeLocation(2, 6), gts.Seq("gcat"), ints(2, 3, 4, 5)},
		{gts.NewRangeLocation(4, 8), gts.Seq("atgc"), ints(4, 5, 6, 7)},

		{gts.NewAmbiguousLocation(0, 4), gts.Seq("atgc"), ints(0, 1, 2, 3)},
		{gts.NewAmbiguousLocation(2, 6), gts.Seq("gcat"), ints(2, 3, 4, 5)},
		{gts.NewAmbiguousLocation(4, 8), gts.Seq("atgc"), ints(4, 5, 6, 7)},

		{gts.NewBetweenLocation(0, 4), gts.Seq("atgc"), ints(0, 1, 2, 3)},
		{gts.NewBetweenLocation(2, 6), gts.Seq("gcat"), ints(2, 3, 4, 5)},
		{gts.NewBetweenLocation(4, 8), gts.Seq("atgc"), ints(4, 5, 6, 7)},

		{
			gts.NewComplementLocation(gts.NewRangeLocation(0, 4)),
			gts.Complement(gts.Seq("atgc")), ints(0, 1, 2, 3),
		},
		{
			gts.NewComplementLocation(gts.NewRangeLocation(2, 6)),
			gts.Complement(gts.Seq("gcat")), ints(2, 3, 4, 5),
		},
		{
			gts.NewComplementLocation(gts.NewRangeLocation(4, 8)),
			gts.Complement(gts.Seq("atgc")), ints(4, 5, 6, 7),
		},

		{
			gts.NewJoinLocation([]gts.Location{
				gts.NewRangeLocation(0, 2),
				gts.NewRangeLocation(3, 5),
				gts.NewRangeLocation(6, 8),
			}),
			gts.Seq("atcagc"),
			ints(0, 1, 3, 4, 6, 7),
		},
		{
			gts.NewJoinLocation([]gts.Location{
				gts.NewRangeLocation(0, 3),
				gts.NewRangeLocation(5, 8),
			}),
			gts.Seq("atgtgc"),
			ints(0, 1, 2, 5, 6, 7),
		},

		{
			gts.NewOrderLocation([]gts.Location{
				gts.NewRangeLocation(0, 2),
				gts.NewRangeLocation(3, 5),
				gts.NewRangeLocation(6, 8),
			}),
			gts.Seq("atcagc"),
			ints(0, 1, 3, 4, 6, 7),
		},
		{
			gts.NewOrderLocation([]gts.Location{
				gts.NewRangeLocation(0, 3),
				gts.NewRangeLocation(5, 8),
			}),
			gts.Seq("atgtgc"),
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

func TestLocationsMap(t *testing.T) {
	values := []struct {
		Locstr string
		Expect []int
	}{
		{"1", ints(0)},
		{"1..2", ints(0, 1)},
		{"1.2", ints(0, 1)},
		{"1^2", ints(0, 1)},
		{"complement(1..2)", ints(0, 1)},
		{"join(1..2,42..43)", ints(0, 1, 41, 42)},
		{"order(1..2,42..43)", ints(0, 1, 41, 42)},
	}

	cases := make([]assert.F, len(values))
	for i, value := range values {
		loc := forceLocation(value.Locstr)
		idx := make([]int, loc.Len())
		for i := range idx {
			idx[i] = loc.Map(i)
		}
		cases[i] = assert.All(
			assert.Equal(idx, value.Expect),
			assert.Panic(func() { loc.Map(-1) }),
			assert.Panic(func() { loc.Map(loc.Len() + 1) }),
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
	"PointLocationParser":      gts.PointLocationParser,
	"RangeLocationParser":      gts.RangeLocationParser,
	"AmbiguousLocationParser":  gts.AmbiguousLocationParser,
	"BetweenLocationParser":    gts.BetweenLocationParser,
	"OrderLocationParser":      gts.OrderLocationParser,
	"JoinLocationParser":       gts.JoinLocationParser,
	"ComplementLocationParser": gts.ComplementLocationParser,
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
				assert.Equal(r.Value.(gts.Location).String(), e),
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

	for _, pair := range locationStrings {
		_, err := gts.AsLocation(pair.Value)
		cases = append(cases, assert.NoError(err))
	}

	_, err := gts.AsLocation("")
	cases = append(cases, assert.IsError(err))

	assert.Apply(t, cases...)
}
