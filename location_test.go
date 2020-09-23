package gts

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/go-gts/gts/testutils"
	"github.com/go-pars/pars"
	"github.com/go-test/deep"
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
		t.Run(tt.in.String(), func(t *testing.T) {
			if tt.in.String() != tt.str {
				t.Errorf("%s.String() = %q, want %q", locRep(tt.in), tt.in.String(), tt.str)
			}
			if tt.in.Len() != tt.len {
				t.Errorf("%s.Len() = %d, want %d", locRep(tt.in), tt.in.Len(), tt.len)
			}
			in := tt.in.Complement()
			if in.Len() != tt.len {
				t.Errorf("%s.Len() = %d, want %d", locRep(in), in.Len(), tt.len)
			}
		})
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
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Range(0, 2), Point(0), 0, -1},
	{Range(0, 2), Range(0, 1), 0, -1},
	{Range(0, 2), Between(0), 0, -2},
	{Range(1, 3), Range(0, 2), 0, -1},
	{Range(0, 2), Range(0, 2), 2, 1},
	{Range(0, 2), Range(0, 2), 2, -1},
	{PartialRange(0, 4, PartialBoth), Join(PartialRange(0, 2, Partial5), PartialRange(3, 5, Partial3)), 2, 1},

	{Join(Range(0, 2), Range(3, 5)), Join(Range(1, 3), Range(4, 6)), 0, 1},
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Join(Range(0, 2), Range(3, 5)), Join(Point(0), Range(2, 4)), 0, -1},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 1), Range(2, 4)), 0, -1},
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
	{Ambiguous{0, 4}, Order(Ambiguous{0, 2}, Ambiguous{3, 5}), 2, 1},

	{Order(Range(0, 2), Range(3, 5)), Order(Range(1, 3), Range(4, 6)), 0, 1},
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Order(Range(0, 2), Range(3, 5)), Order(Point(0), Range(2, 4)), 0, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 1), Range(2, 4)), 0, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(4, 6)), 2, 1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(2, 4)), 2, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, 1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, -1},
}

func TestLocationShift(t *testing.T) {
	for i, tt := range locationShiftTests {
		if !reflect.DeepEqual(shiftLocation(tt.in, tt.i, tt.n), tt.out) {
			t.Errorf(
				"\ncase [%d]:\nshift by (%d, %d)\n in: %s\nout: %s\nexp: %s",
				i+1, tt.i, tt.n,
				locRep(tt.in),
				locRep(shiftLocation(tt.in, tt.i, tt.n)),
				locRep(tt.out),
			)
		}
		if !reflect.DeepEqual(
			shiftLocation(tt.in.Complement(), tt.i, tt.n),
			tt.out.Complement(),
		) {
			t.Errorf(
				"\ncase [%d]:\nshift by (%d, %d)\n in: %s\nout: %s\nexp: %s",
				i+1, tt.i, tt.n,
				locRep(tt.in.Complement()),
				locRep(shiftLocation(tt.in.Complement(), tt.i, tt.n)),
				locRep(tt.out.Complement()),
			)
		}
	}
}

var locationExpandTests = []struct {
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
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Range(0, 2), Point(0), 0, -1},
	{Range(0, 2), Range(0, 1), 0, -1},
	{Range(0, 2), Between(0), 0, -2},
	{Range(1, 3), Range(0, 2), 0, -1},
	{Range(0, 2), Range(0, 2), 2, 1},
	{Range(0, 2), Range(0, 2), 2, -1},
	{PartialRange(0, 4, PartialBoth), PartialRange(0, 5, PartialBoth), 2, 1},

	{Join(Range(0, 2), Range(3, 5)), Join(Range(1, 3), Range(4, 6)), 0, 1},
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Join(Range(0, 2), Range(3, 5)), Join(Point(0), Range(2, 4)), 0, -1},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 1), Range(2, 4)), 0, -1},
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
	{Ambiguous{0, 4}, Ambiguous{0, 5}, 2, 1},

	{Order(Range(0, 2), Range(3, 5)), Order(Range(1, 3), Range(4, 6)), 0, 1},
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Order(Range(0, 2), Range(3, 5)), Order(Point(0), Range(2, 4)), 0, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 1), Range(2, 4)), 0, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(4, 6)), 2, 1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(2, 4)), 2, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, 1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, -1},
}

func TestLocationExpand(t *testing.T) {
	for i, tt := range locationExpandTests {
		if !reflect.DeepEqual(tt.in.Expand(tt.i, tt.n), tt.out) {
			t.Errorf(
				"\ncase [%d]:\nexpand by (%d, %d)\n in: %s\nout: %s\nexp: %s",
				i+1, tt.i, tt.n,
				locRep(tt.in),
				locRep(tt.in.Expand(tt.i, tt.n)),
				locRep(tt.out),
			)
		}
		if !reflect.DeepEqual(
			tt.in.Complement().Expand(tt.i, tt.n),
			tt.out.Complement(),
		) {
			t.Errorf(
				"\ncase [%d]:\nexpand by (%d, %d)\n in: %s\nout: %s\nexp: %s",
				i+1, tt.i, tt.n,
				locRep(tt.in.Complement()),
				locRep(tt.in.Complement().Expand(tt.i, tt.n)),
				locRep(tt.out.Complement()),
			)
		}
	}
}

type NullLocation int

func (null NullLocation) String() string {
	return "nil"
}

func (null NullLocation) Len() int {
	return 0
}

func (null NullLocation) Shift(i, n int) Location {
	return null
}

func (null NullLocation) Expand(i, n int) Location {
	return null
}

func (null NullLocation) Less(loc Location) bool {
	return false
}

func (null NullLocation) Complement() Location {
	return Complemented{null}
}

func (null NullLocation) Locate(seq Sequence) Sequence {
	return WithBytes(seq, nil)
}

var locationLessTests = []struct {
	loc  Location
	pass []Location
	fail []Location
}{
	{
		Between(1),
		[]Location{
			Between(2),
			Point(1),
			Range(1, 3),
			Join(Range(1, 3), Range(4, 6)),
			Ambiguous{1, 3},
			Order(Range(1, 3), Range(4, 6)),
			NullLocation(0),
		},
		[]Location{
			Between(1),
			Point(0),
			Range(0, 2),
			Join(Range(0, 2), Range(3, 5)),
			Ambiguous{0, 2},
			Order(Range(0, 2), Range(3, 5)),
		},
	},
	{
		Point(0),
		[]Location{
			Between(1),
			Point(1),
			Range(1, 3),
			Join(Range(1, 3), Range(4, 6)),
			Ambiguous{1, 3},
			Order(Range(1, 3), Range(4, 6)),
			NullLocation(0),
		},
		[]Location{
			Between(0),
			Point(0),
			Range(0, 2),
			Join(Range(0, 2), Range(3, 5)),
			Ambiguous{0, 2},
			Order(Range(0, 2), Range(3, 5)),
		},
	},
	{
		Range(1, 3),
		[]Location{
			Between(2),
			Point(1),
			Range(2, 3),
			Range(1, 4),
			Join(Range(2, 4), Range(5, 7)),
			Ambiguous{2, 3},
			Ambiguous{1, 4},
			Order(Range(2, 4), Range(5, 7)),
			NullLocation(0),
		},
		[]Location{
			Between(1),
			Point(0),
			Range(0, 3),
			Range(1, 2),
			Range(1, 3),
			Join(Range(1, 3), Range(4, 6)),
			Ambiguous{0, 3},
			Ambiguous{1, 2},
			Ambiguous{1, 3},
			Order(Range(1, 3), Range(4, 6)),
		},
	},
	{PartialRange(1, 3, Partial5), []Location{Range(1, 3), Ambiguous{1, 3}}, nil},
	{PartialRange(1, 3, Partial3), nil, []Location{Range(1, 3), Ambiguous{1, 3}}},
	{Range(1, 3), nil, []Location{PartialRange(1, 3, Partial5)}},
	{Range(1, 3), []Location{PartialRange(1, 3, Partial3)}, nil},
	{
		Join(Range(0, 2), Range(3, 5)),
		[]Location{Join(Range(1, 2), Range(3, 5))},
		[]Location{Join(Range(0, 2), Range(3, 5))},
	},
	{
		Ambiguous{1, 3},
		[]Location{
			Between(2),
			Point(1),
			Range(2, 3),
			Range(1, 4),
			PartialRange(1, 3, Partial3),
			Join(Range(2, 4), Range(5, 7)),
			Ambiguous{2, 3},
			Ambiguous{1, 4},
			Order(Range(2, 4), Range(5, 7)),
			NullLocation(0),
		},
		[]Location{
			Between(1),
			Point(0),
			Range(0, 3),
			PartialRange(1, 3, Partial5),
			Range(1, 2),
			Range(1, 3),
			Join(Range(1, 3), Range(4, 6)),
			Ambiguous{0, 3},
			Ambiguous{1, 3},
			Order(Range(1, 3), Range(4, 6)),
		},
	},
	{
		Order(Range(0, 2), Range(3, 5)),
		[]Location{Order(Range(1, 2), Range(3, 5))},
		[]Location{Order(Range(0, 2), Range(3, 5))},
	},
}

func locationLessPassTest(t *testing.T, lhs, rhs Location) {
	if !lhs.Less(rhs) {
		t.Errorf("expected %s < %s", locRep(lhs), locRep(rhs))
	}
	if _, ok := lhs.(Complemented); !ok {
		locationLessPassTest(t, lhs.Complement(), rhs)
	}
	if _, ok := rhs.(Complemented); !ok {
		locationLessPassTest(t, lhs, rhs.Complement())
	}
}

func locationLessFailTest(t *testing.T, lhs, rhs Location) {
	if lhs.Less(rhs) {
		t.Errorf("expected %s >= %s", locRep(lhs), locRep(rhs))
	}
	if _, ok := lhs.(Complemented); !ok {
		locationLessFailTest(t, lhs.Complement(), rhs)
	}
	if _, ok := rhs.(Complemented); !ok {
		locationLessFailTest(t, lhs, rhs.Complement())
	}
}

func TestLocationLess(t *testing.T) {
	for _, tt := range locationLessTests {
		for _, loc := range tt.pass {
			locationLessPassTest(t, tt.loc, loc)
		}
		for _, loc := range tt.fail {
			locationLessFailTest(t, tt.loc, loc)
		}
	}
}

var locationFlipTest = []struct {
	in  Location
	out Location
}{
	{NullLocation(0), NullLocation(0)},
	{Between(0), Between(9)},
	{Point(0), Point(9)},
	{Range(0, 3), Range(7, 10)},
	{PartialRange(0, 3, Partial5), PartialRange(7, 10, Partial3)},
	{PartialRange(0, 3, Partial3), PartialRange(7, 10, Partial5)},
	{PartialRange(0, 3, PartialBoth), PartialRange(7, 10, PartialBoth)},
	{Join(Range(0, 3), Range(5, 8)), Join(Range(2, 5), Range(7, 10))},
	{Range(0, 3).Complement(), Range(7, 10).Complement()},
	{Ambiguous{0, 3}, Ambiguous{7, 10}},
	{Order(Range(0, 3), Range(5, 8)), Order(Range(2, 5), Range(7, 10))},
}

func TestLocationFlip(t *testing.T) {
	for _, tt := range locationFlipTest {
		out := flipLocation(tt.in, 10)
		testutils.Equals(t, out, tt.out)
	}
}

var locationNormalizeTest = []struct {
	in  Location
	out Location
}{
	{NullLocation(0), NullLocation(0)},
	{Between(10), Between(0)},
	{Point(10), Point(0)},
	{Range(10, 13), Range(0, 3)},
	{Range(8, 12), Join(Range(8, 10), Range(0, 2))},
	{PartialRange(8, 12, PartialBoth), Join(PartialRange(8, 10, Partial5), PartialRange(0, 2, Partial3))},
	{Join(Range(10, 13), Range(5, 8)), Join(Range(0, 3), Range(5, 8))},
	{Range(10, 13).Complement(), Range(0, 3).Complement()},
	{Ambiguous{10, 13}, Ambiguous{0, 3}},
	{Order(Range(10, 13), Range(5, 8)), Order(Range(0, 3), Range(5, 8))},
}

func TestLocationNormalize(t *testing.T) {
	for _, tt := range locationNormalizeTest {
		out := normalizeLocation(tt.in, 10)
		testutils.Equals(t, out, tt.out)
	}
}

var locationReductionTests = []struct {
	in  Location
	out Location
}{
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Range(0, 1), Point(0)},
	{Join(Point(0), Point(0)), Point(0)},
	{Join(Point(0), Range(0, 2)), Range(0, 2)},
	{Join(Range(0, 2), Point(2)), Range(0, 2)},
	{Join(Range(0, 2), Range(2, 4)), Range(0, 4)},
	{Join(Range(2, 4).Complement(), Range(0, 2).Complement()), Range(0, 4).Complement()},
	{Order(Range(0, 2)), Range(0, 2)},
}

func TestLocationReduction(t *testing.T) {
	for _, tt := range locationReductionTests {
		testutils.Equals(t, tt.in, tt.out)
	}
}

var locateTests = []struct {
	in  Location
	out Sequence
}{
	{Between(0), New(nil, nil, []byte(""))},
	{Point(0), New(nil, nil, []byte("a"))},
	{Range(0, 2), New(nil, nil, []byte("at"))},
	{Ambiguous{0, 2}, New(nil, nil, []byte("at"))},
	{Join(Range(0, 2), Range(3, 5)), New(nil, nil, []byte("atca"))},
	{Order(Range(0, 2), Range(3, 5)), New(nil, nil, []byte("atca"))},
}

func TestLocate(t *testing.T) {
	seq := New(nil, nil, []byte("atgcatgc"))
	for _, tt := range locateTests {
		out, exp := tt.in.Locate(seq), tt.out
		if !reflect.DeepEqual(out.Info(), exp.Info()) {
			t.Errorf("Slice(in, %d, %d).Info() = %v, want %v", 2, 6, out.Info(), exp.Info())
		}
		if diff := deep.Equal(out.Features(), exp.Features()); diff != nil {
			t.Errorf("Slice(in, %d, %d).Features() = %v, want %v", 2, 6, out.Features(), exp.Features())
		}
		if diff := deep.Equal(out.Bytes(), exp.Bytes()); diff != nil {
			t.Errorf("Slice(in, %d, %d).Bytes() = %v, want %v", 2, 6, out.Bytes(), exp.Bytes())
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

var locationParserPassTests = []struct {
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

var locationParserFailTests = []struct {
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
	for _, tt := range locationParserPassTests {
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

	for _, tt := range locationParserFailTests {
		prs := pars.Exact(tt.prs)
		_, err := prs.Parse(pars.FromString(tt.in))
		if err == nil {
			t.Errorf("expected error while parsing %q", tt.in)
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
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("parsed %q: expected %s, got %s", tt.in, locRep(tt.out), locRep(out))
		}
	}
}

func TestLocationPanics(t *testing.T) {
	testutils.Panics(t, func() { Range(2, 0) })
	testutils.Panics(t, func() { Join() })
	testutils.Panics(t, func() { Order() })
}
