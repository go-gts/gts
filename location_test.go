package gts

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	pars "gopkg.in/pars.v2"
)

func locRep(loc Location) string {
	switch v := loc.(type) {
	case Between:
		return fmt.Sprintf("Between(%d)", v)
	case Point:
		return fmt.Sprintf("Point(%d)", v)
	case Ranged:
		return fmt.Sprintf("Ranged(%d, %d, %v)", v.Start, v.End, v.Partial)
	case Complemented:
		return fmt.Sprintf("Complemented(%s)", locRep(v[0]))
	case Joined:
		ss := make([]string, len(v))
		for i, u := range v {
			ss[i] = locRep(u)
		}
		return fmt.Sprintf("Joined(%s)", strings.Join(ss, ", "))
	case Ambiguous:
		return fmt.Sprintf("Ambiguous(%d, %d)", v[0], v[1])
	case Ordered:
		ss := make([]string, len(v))
		for i, u := range v {
			ss[i] = locRep(u)
		}
		return fmt.Sprintf("Ordered(%s)", strings.Join(ss, ", "))
	default:
		return "Unknown"
	}
}

var locationAccessorTests = []struct {
	in  Location
	str string
	len int
}{
	{Between(0), "0^1", 0},
	{Point(0), "1", 1},
	{Range(0, 2), "1..2", 2},

	{PartialRange(0, 2, Complete), "1..2", 2},
	{PartialRange(0, 2, Partial5), "<1..2", 2},
	{PartialRange(0, 2, Partial3), "1..>2", 2},
	{PartialRange(0, 2, PartialBoth), "<1..>2", 2},

	{Join(Range(0, 2), Range(3, 5)), "join(1..2,4..5)", 4},
	{Join(Range(0, 2), Join(Range(3, 5), Range(6, 8))), "join(1..2,4..5,7..8)", 6},
	{Join(Point(0), Point(2)), "join(1,3)", 2},

	{Ambiguous{0, 2}, "1.2", 1},

	{Order(Range(0, 2), Range(2, 4)), "order(1..2,3..4)", 4},
	{Order(Range(0, 2), Order(Range(2, 4), Range(4, 6))), "order(1..2,3..4,5..6)", 6},
	{Order(Point(0), Point(2)), "order(1,3)", 2},
}

func TestLocationAccessors(t *testing.T) {
	for _, tt := range locationAccessorTests {
		if tt.in.String() != tt.str {
			t.Errorf("%s.String() = %q, want %q", locRep(tt.in), tt.in.String(), tt.str)
		}
		if tt.in.Len() != tt.len {
			t.Errorf("%s.Len() = %d, want %d", locRep(tt.in), tt.in.Len(), tt.len)
		}
	}
}

var locationShiftTests = []struct {
	in, out Location
	i, n    int
}{
	{Between(0), Between(0), 0, 1},
	{Between(0), Between(0), 0, -1},
	{Between(1), Between(2), 0, 1},
	{Between(1), Between(0), 0, -1},

	{Point(0), Point(1), 0, 1},
	{Point(0), Between(0), 0, -1},
	{Point(0), Point(0), 1, 1},
	{Point(0), Point(0), 1, -1},
	{Point(1), Point(2), 0, 1},
	{Point(1), Point(0), 0, -1},

	{Range(0, 2), Range(1, 3), 0, 1},
	{Range(0, 2), Point(0), 0, -1},
	{Range(0, 2), Between(0), 0, -2},
	{Range(1, 3), Range(0, 2), 0, -1},
	{Range(0, 2), Range(0, 2), 2, 1},
	{Range(0, 2), Range(0, 2), 2, -1},

	{Join(Range(0, 2), Range(3, 5)), Join(Range(1, 3), Range(4, 6)), 0, 1},
	{Join(Range(0, 2), Range(3, 5)), Join(Point(0), Range(2, 4)), 0, -1},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 2), Range(4, 6)), 2, 1},
	{Join(Range(0, 2), Range(3, 5)), Range(0, 4), 2, -1},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 2), Range(3, 5)), 5, 1},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 2), Range(3, 5)), 5, -1},

	{Ambiguous{0, 2}, Ambiguous{1, 3}, 0, 1},
	{Ambiguous{0, 2}, Ambiguous{0, 1}, 0, -1},
	{Ambiguous{0, 2}, Between(0), 0, -2},
	{Ambiguous{1, 3}, Ambiguous{0, 2}, 0, -1},
	{Ambiguous{0, 2}, Ambiguous{0, 2}, 2, 1},
	{Ambiguous{0, 2}, Ambiguous{0, 2}, 2, -1},

	{Order(Range(0, 2), Range(3, 5)), Order(Range(1, 3), Range(4, 6)), 0, 1},
	{Order(Range(0, 2), Range(3, 5)), Order(Point(0), Range(2, 4)), 0, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(4, 6)), 2, 1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(2, 4)), 2, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, 1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, -1},
}

func areLocatable(locs ...Location) bool {
	for _, loc := range locs {
		if _, ok := loc.(Locatable); !ok {
			return false
		}
	}
	return true
}

func TestLocationShift(t *testing.T) {
	for _, tt := range locationShiftTests {
		if !reflect.DeepEqual(tt.in.Shift(tt.i, tt.n), tt.out) {
			t.Errorf(
				"%s.Shift(%d, %d) = %s, want %s",
				locRep(tt.in), tt.i, tt.n,
				locRep(tt.in.Shift(tt.i, tt.n)),
				locRep(tt.out),
			)
		}
		if areLocatable(tt.in, tt.out) {
			if !reflect.DeepEqual(
				tt.in.(Locatable).Complement().Shift(tt.i, tt.n),
				tt.out.(Locatable).Complement(),
			) {
				t.Errorf(
					"%s.Shift(%d, %d) = %s, want %s",
					locRep(tt.in.(Locatable).Complement()), tt.i, tt.n,
					locRep(tt.in.(Locatable).Complement().Shift(tt.i, tt.n)),
					locRep(tt.out.(Locatable).Complement()),
				)
			}
		}
	}
}

var locationReductionTests = []struct {
	in  Location
	out Location
}{
	{Range(0, 1), Point(0)},
	{Join(Point(0), Point(0)), Point(0)},
	{Join(Point(0), Range(0, 2)), Range(0, 2)},
	{Join(Range(0, 2), Point(2)), Range(0, 2)},
	{Join(Range(0, 2), Range(2, 4)), Range(0, 4)},
	{Join(Range(2, 4).Complement(), Range(0, 2).Complement()), Range(0, 4).Complement()},
	{Order(Range(0, 2)), Range(0, 2)},
}

func TestLocationReduction(t *testing.T) {
	for _, tt := range locationReductionTests {
		equals(t, tt.in, tt.out)
	}
}

var locatableTests = []struct {
	in  Locatable
	out Sequence
}{
	{Between(0), New(nil, []byte(""))},
	{Point(0), New(nil, []byte("a"))},
	{Range(0, 2), New(nil, []byte("at"))},
	{Join(Range(0, 2), Range(3, 5)), New(nil, []byte("atca"))},
}

func TestLocatable(t *testing.T) {
	seq := New(nil, []byte("atgcatgc"))
	for _, tt := range locatableTests {
		out := tt.in.Locate(seq)
		exp := tt.out
		if !Equal(out, exp) {
			t.Errorf(
				"%s.Locate(%q) = %q, want %q",
				locRep(tt.in), string(seq.Bytes()),
				string(out.Bytes()), string(exp.Bytes()),
			)
		}
		cmp := tt.in.Complement()
		cmpstr := fmt.Sprintf("complement(%s)", tt.in)
		if cmp.String() != cmpstr {
			t.Errorf("%s.String() = %q, want %q", locRep(cmp), cmp, cmpstr)
		}
		if cmp.Len() != tt.in.Len() {
			t.Errorf("%s.Len() = %d, want %d", locRep(cmp), cmp.Len(), tt.in.Len())
		}
		if cmp.Complement().String() != tt.in.String() {
			t.Errorf(
				"%s.Complement() = %s, want %s",
				locRep(cmp), locRep(cmp.Complement()), locRep(tt.in),
			)
		}
		out = cmp.Locate(seq)
		exp = Reverse(Complement(tt.out))
		if !Equal(out, exp) {
			t.Errorf(
				"%s.Locate(%q) = %q, want %q",
				locRep(cmp), string(seq.Bytes()),
				string(out.Bytes()), string(exp.Bytes()),
			)
		}
	}
}

var locationParserSuccessTests = []struct {
	prs pars.Parser
	loc Location
}{
	{BetweenParser, Between(0)},
	{PointParser, Point(0)},
	{RangeParser, Range(0, 2)},
	{RangeParser, PartialRange(0, 2, Partial5)},
	{RangeParser, PartialRange(0, 2, Partial3)},
	{RangeParser, PartialRange(0, 2, PartialBoth)},
	{RangeParser, PartialRange(0, 2, Partial3)},
	{RangeParser, PartialRange(0, 2, PartialBoth)},
	{ComplementParser, Range(0, 2).Complement()},
	{JoinParser, Join(Range(0, 2), Range(3, 5))},
	{AmbiguousParser, Ambiguous{0, 2}},
	{OrderParser, Order(Range(0, 2), Range(2, 4))},
}

var locationParserFailureTests = []struct {
	prs pars.Parser
	in  string
}{
	{BetweenParser, ""},
	{BetweenParser, "?"},
	{BetweenParser, "1"},
	{BetweenParser, "1?"},
	{BetweenParser, "1^?"},
	{BetweenParser, "1^3"},

	{PointParser, ""},
	{PointParser, "?"},

	{RangeParser, ""},
	{RangeParser, "?"},
	{RangeParser, "1"},
	{RangeParser, "1??"},
	{RangeParser, "1..?"},

	{ComplementParser, ""},
	{ComplementParser, "complement?"},
	{ComplementParser, "complement(?"},
	{ComplementParser, "complement(1..2"},
	{ComplementParser, "complement(1..2?"},

	{JoinParser, ""},
	{JoinParser, "join?"},
	{JoinParser, "join("},
	{JoinParser, "join(1..2,?"},
	{JoinParser, "join(1..2,3..5"},
	{JoinParser, "join(1..2,3..5?"},

	{OrderParser, ""},
	{OrderParser, "order?"},
	{OrderParser, "order("},
	{OrderParser, "order(1..2,?"},
	{OrderParser, "order(1..2,3..5"},
	{OrderParser, "order(1..2,3..5?"},

	{AmbiguousParser, ""},
	{AmbiguousParser, "?"},
	{AmbiguousParser, "1"},
	{AmbiguousParser, "1?"},
	{AmbiguousParser, "1.?"},
}

func TestLocationParsers(t *testing.T) {
	for _, tt := range locationParserSuccessTests {
		prs := pars.Exact(tt.prs)
		in := tt.loc.String()
		res, err := prs.Parse(pars.FromString(in))
		if err != nil {
			t.Errorf("while parsing %q got: %v", in, err)
			continue
		}
		out, ok := res.Value.(Location)
		if !ok {
			t.Errorf("parsed result is of type `%T`, want Location", res.Value)
			continue
		}
		if !reflect.DeepEqual(out, tt.loc) {
			t.Errorf("parser output is %s, want %s", locRep(out), locRep(tt.loc))
		}
	}

	for _, tt := range locationParserFailureTests {
		prs := pars.Exact(tt.prs)
		_, err := prs.Parse(pars.FromString(tt.in))
		if err == nil {
			t.Errorf("expected error while parsing %q", tt.in)
		}
	}
}

var locatableParserTests = []struct {
	in  string
	out Locatable
}{
	{"0^1", Between(0)},
	{"1", Point(0)},
	{"1..2", Range(0, 2)},
	{"<1..2", PartialRange(0, 2, Partial5)},
	{"1..>2", PartialRange(0, 2, Partial3)},
	{"<1..>2", PartialRange(0, 2, PartialBoth)},
	{"1..2>", PartialRange(0, 2, Partial3)},
	{"<1..2>", PartialRange(0, 2, PartialBoth)},
	{"join(1..2,4..5)", Join(Range(0, 2), Range(3, 5))},
	{"join(1..2, 4..5)", Join(Range(0, 2), Range(3, 5))},
}

func TestLocatableParser(t *testing.T) {
	for _, tt := range locatableParserTests {
		res, err := LocatableParser.Parse(pars.FromString(tt.in))
		if err != nil {
			t.Errorf("failed to parse %q: %v", tt.in, err)
			continue
		}
		out, ok := res.Value.(Locatable)
		if !ok {
			t.Errorf("parsed result is of type `%T`, want Locatable", res.Value)
			continue
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("parsed %q: expected %s, got %s", tt.in, locRep(tt.out), locRep(out))
		}
	}
}

var locationParserTests = []struct {
	in  string
	out Location
}{
	{"0^1", Between(0)},
	{"1", Point(0)},
	{"1..2", Range(0, 2)},
	{"<1..2", PartialRange(0, 2, Partial5)},
	{"1..>2", PartialRange(0, 2, Partial3)},
	{"<1..>2", PartialRange(0, 2, PartialBoth)},
	{"1..2>", PartialRange(0, 2, Partial3)},
	{"<1..2>", PartialRange(0, 2, PartialBoth)},
	{"join(1..2,4..5)", Join(Range(0, 2), Range(3, 5))},
	{"1.2", Ambiguous{0, 2}},
	{"order(1..2,3..4)", Order(Range(0, 2), Range(2, 4))},
	{"order(1..2, 3..4)", Order(Range(0, 2), Range(2, 4))},
}

func TestLocationParser(t *testing.T) {
	for _, tt := range locationParserTests {
		res, err := LocationParser.Parse(pars.FromString(tt.in))
		if err != nil {
			t.Errorf("failed to parse %q: %v", tt.in, err)
			continue
		}
		out, ok := res.Value.(Location)
		if !ok {
			t.Errorf("parsed result is of type `%T`, want Location", res.Value)
			return
			continue
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("parsed %q: expected %s, got %s", tt.in, locRep(tt.out), locRep(out))
		}
	}
}

func TestLocationPanics(t *testing.T) {
	panics(t, func() { Range(2, 0) })
	panics(t, func() { Join() })
	panics(t, func() { Order() })
}
