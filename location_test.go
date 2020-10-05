package gts

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-pars/pars"
)

type NullLocation int

func (null NullLocation) String() string {
	return "nil"
}

func (null NullLocation) Len() int {
	return 0
}

func (null NullLocation) Head() int {
	return int(null)
}

func (null NullLocation) Tail() int {
	return int(null)
}

func (null NullLocation) Less(loc Location) bool {
	return false
}

func (null NullLocation) Region() Region {
	return Segment{}
}

func (null NullLocation) Complement() Location {
	return Complemented{null}
}

func (null NullLocation) Reverse(length int) Location {
	return null
}

func (null NullLocation) Normalize(length int) Location {
	return null
}

func (null NullLocation) Shift(i, n int) Location {
	return null
}

func (null NullLocation) Expand(i, n int) Location {
	return null
}

func locRep(loc Location) string {
	switch v := loc.(type) {
	case Between:
		return fmt.Sprintf("Between(%d)", v)
	case Point:
		return fmt.Sprintf("Point(%d)", v)
	case Ranged:
		return fmt.Sprintf("Ranged(%d, %d, %v)", v.Start, v.End, v.Partial)
	case Complemented:
		return fmt.Sprintf("Complemented(%s)", locRep(v.Location))
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
	reg Region
}{
	{Between(0), "0^1", 0, Segment{0, 0}},
	{Point(0), "1", 1, Segment{0, 1}},
	{Range(0, 2), "1..2", 2, Segment{0, 2}},

	{PartialRange(0, 2, Complete), "1..2", 2, Segment{0, 2}},
	{PartialRange(0, 2, Partial5), "<1..2", 2, Segment{0, 2}},
	{PartialRange(0, 2, Partial3), "1..>2", 2, Segment{0, 2}},
	{PartialRange(0, 2, PartialBoth), "<1..>2", 2, Segment{0, 2}},

	{Ambiguous{0, 2}, "1.2", 1, Segment{0, 2}},

	{Join(Range(0, 2), Range(3, 5)), "join(1..2,4..5)", 4, Regions{Segment{0, 2}, Segment{3, 5}}},
	{Join(Range(0, 2), Join(Range(3, 5), Range(6, 8))), "join(1..2,4..5,7..8)", 6, Regions{Segment{0, 2}, Segment{3, 5}, Segment{6, 8}}},
	{Join(Point(0), Point(2)), "join(1,3)", 2, Regions{Segment{0, 1}, Segment{2, 3}}},

	{Order(Range(0, 2), Range(2, 4)), "order(1..2,3..4)", 4, Regions{Segment{0, 2}, Segment{2, 4}}},
	{Order(Range(0, 2), Order(Range(2, 4), Range(4, 6))), "order(1..2,3..4,5..6)", 6, Regions{Segment{0, 2}, Segment{2, 4}, Segment{4, 6}}},
	{Order(Point(0), Point(2)), "order(1,3)", 2, Regions{Segment{0, 1}, Segment{2, 3}}},

	{Range(0, 2).Complement(), "complement(1..2)", 2, Segment{2, 0}},
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
			if !reflect.DeepEqual(tt.in.Region(), tt.reg) {
				t.Errorf("%s.Region() = %v, want %v", locRep(tt.in), tt.in.Region(), tt.reg)
			}
			in := tt.in.Complement()
			if in.Len() != tt.len {
				t.Errorf("%s.Len() = %d, want %d", locRep(in), in.Len(), tt.len)
			}
			if !reflect.DeepEqual(in.Region(), tt.reg.Complement()) {
				t.Errorf("%s.Region() = %v, want %v", locRep(in), in.Region(), tt.reg.Complement())
			}
		})
	}
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

var locationReverseTest = []struct {
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

func TestLocationReverse(t *testing.T) {
	for _, tt := range locationReverseTest {
		out := tt.in.Reverse(10)
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
		out := tt.in.Normalize(10)
		testutils.Equals(t, out, tt.out)
	}
}

var locationShiftTests = []struct {
	in, out Location
	i, n    int
}{
	{Between(3), Between(3), 2, 0},
	{Between(3), Between(3), 3, 0},
	{Between(3), Between(3), 4, 0},
	{Between(3), Between(4), 2, 1},
	{Between(3), Between(3), 3, 1},
	{Between(3), Between(3), 4, 1},
	{Between(3), Between(2), 2, -1},
	{Between(3), Between(3), 3, -1},
	{Between(3), Between(3), 4, -1},

	{Point(3), Point(3), 2, 0},
	{Point(3), Point(3), 3, 0},
	{Point(3), Point(3), 4, 0},
	{Point(3), Point(4), 2, 1},
	{Point(3), Point(4), 3, 1},
	{Point(3), Point(3), 4, 1},
	{Point(3), Point(2), 2, -1},
	{Point(3), Between(3), 3, -1},
	{Point(3), Point(3), 4, -1},

	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 2, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 3, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 4, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 5, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 6, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 7, 0},

	{Ranged{3, 6, Complete}, Ranged{4, 7, Complete}, 2, 1},
	{Ranged{3, 6, Complete}, Ranged{4, 7, Complete}, 3, 1},
	{Ranged{3, 6, Complete}, Join(Ranged{3, 4, Complete}, Ranged{5, 7, Complete}), 4, 1},
	{Ranged{3, 6, Partial5}, Join(Ranged{3, 4, Partial5}, Ranged{5, 7, Complete}), 4, 1},
	{Ranged{3, 6, Partial3}, Join(Ranged{3, 4, Complete}, Ranged{5, 7, Partial3}), 4, 1},
	{Ranged{3, 6, PartialBoth}, Join(Ranged{3, 4, Partial5}, Ranged{5, 7, Partial3}), 4, 1},
	{Ranged{3, 6, Complete}, Join(Ranged{3, 5, Complete}, Ranged{6, 7, Complete}), 5, 1},
	{Ranged{3, 6, Partial5}, Join(Ranged{3, 5, Partial5}, Ranged{6, 7, Complete}), 5, 1},
	{Ranged{3, 6, Partial3}, Join(Ranged{3, 5, Complete}, Ranged{6, 7, Partial3}), 5, 1},
	{Ranged{3, 6, PartialBoth}, Join(Ranged{3, 5, Partial5}, Ranged{6, 7, Partial3}), 5, 1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 6, 1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 7, 1},

	{Ranged{3, 6, Complete}, Ranged{2, 5, Complete}, 2, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 5, Partial5}, 3, -1},
	{Ranged{3, 6, Partial3}, Ranged{3, 5, PartialBoth}, 3, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 5, Complete}, 4, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 5, Partial3}, 5, -1},
	{Ranged{3, 6, Partial5}, Ranged{3, 5, PartialBoth}, 5, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 6, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 7, -1},
	{Ranged{3, 6, Complete}, Between(2), 2, -4},

	{Ambiguous{3, 6}, Ambiguous{3, 6}, 2, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 3, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 4, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 5, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 0},

	{Ambiguous{3, 6}, Ambiguous{4, 7}, 2, 1},
	{Ambiguous{3, 6}, Ambiguous{4, 7}, 3, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1},
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 1},

	{Ambiguous{3, 6}, Ambiguous{2, 5}, 2, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 3, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 3, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 4, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 5, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 5, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, -1},
	{Ambiguous{3, 6}, Between(2), 2, -4},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 2, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 3, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 4, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 5, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 6, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 7, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 12, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 13, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 14, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 15, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 0},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 2, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 3, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 4, Complete}, Ranged{5, 7, Complete}, Ranged{14, 17, Complete}}, 4, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 5, Complete}, Ranged{6, 7, Complete}, Ranged{14, 17, Complete}}, 5, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 6, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 7, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 12, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 13, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 14, Complete}, Ranged{15, 17, Complete}}, 14, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 15, Complete}, Ranged{16, 17, Complete}}, 15, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 1},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{2, 5, Complete}, Ranged{12, 15, Complete}}, 2, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 5, Partial5}, Ranged{12, 15, Complete}}, 3, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 5, Complete}, Ranged{12, 15, Complete}}, 4, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 5, Partial3}, Ranged{12, 15, Complete}}, 5, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 6, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 7, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 12, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 15, Partial5}}, 13, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 15, Complete}}, 14, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 15, Partial3}}, 15, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, -1},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Between(2), Ranged{9, 12, Complete}}, 2, -4},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Between(12)}, 12, -4},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Between(2), 2, -14},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 2, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 3, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 4, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 5, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 6, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 7, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 12, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 13, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 14, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 15, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 0},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 2, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 3, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Joined{Ranged{3, 4, Complete}, Ranged{5, 7, Complete}}, Ranged{14, 17, Complete}}, 4, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Joined{Ranged{3, 5, Complete}, Ranged{6, 7, Complete}}, Ranged{14, 17, Complete}}, 5, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 6, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 7, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 12, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 13, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Joined{Ranged{13, 14, Complete}, Ranged{15, 17, Complete}}}, 14, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Joined{Ranged{13, 15, Complete}, Ranged{16, 17, Complete}}}, 15, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 1},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{2, 5, Complete}, Ranged{12, 15, Complete}}, 2, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 5, Partial5}, Ranged{12, 15, Complete}}, 3, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 5, Complete}, Ranged{12, 15, Complete}}, 4, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 5, Partial3}, Ranged{12, 15, Complete}}, 5, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 6, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 7, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 12, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 15, Partial5}}, 13, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 15, Complete}}, 14, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 15, Partial3}}, 15, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, -1},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Between(2), Ranged{9, 12, Complete}}, 2, -4},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Between(12)}, 12, -4},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Between(2), Between(2)}, 2, -14},
}

func TestLocationShift(t *testing.T) {
	for i, tt := range locationShiftTests {
		if !reflect.DeepEqual(tt.in.Shift(tt.i, tt.n), tt.out) {
			t.Errorf(
				"\ncase [%d]:\nshift by (%d, %d)\n in: %s\nout: %s\nexp: %s",
				i+1, tt.i, tt.n,
				locRep(tt.in),
				locRep(tt.in.Shift(tt.i, tt.n)),
				locRep(tt.out),
			)
		}
		if !reflect.DeepEqual(
			tt.in.Complement().Shift(tt.i, tt.n),
			tt.out.Complement(),
		) {
			t.Errorf(
				"\ncase [%d]:\nshift by (%d, %d)\n in: %s\nout: %s\nexp: %s",
				i+1, tt.i, tt.n,
				locRep(tt.in.Complement()),
				locRep(tt.in.Complement().Shift(tt.i, tt.n)),
				locRep(tt.out.Complement()),
			)
		}
	}
}

var locationExpandTests = []struct {
	in, out Location
	i, n    int
}{
	{Between(3), Between(3), 2, 0},
	{Between(3), Between(3), 3, 0},
	{Between(3), Between(3), 4, 0},
	{Between(3), Between(4), 2, 1},
	{Between(3), Between(3), 3, 1},
	{Between(3), Between(3), 4, 1},
	{Between(3), Between(2), 2, -1},
	{Between(3), Between(3), 3, -1},
	{Between(3), Between(3), 4, -1},

	{Point(3), Point(3), 2, 0},
	{Point(3), Point(3), 3, 0},
	{Point(3), Point(3), 4, 0},
	{Point(3), Point(4), 2, 1},
	{Point(3), Point(4), 3, 1},
	{Point(3), Point(3), 4, 1},
	{Point(3), Point(2), 2, -1},
	{Point(3), Between(3), 3, -1},
	{Point(3), Point(3), 4, -1},

	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 2, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 3, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 4, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 5, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 6, 0},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 7, 0},

	{Ranged{3, 6, Complete}, Ranged{4, 7, Complete}, 2, 1},
	{Ranged{3, 6, Complete}, Ranged{4, 7, Complete}, 3, 1},
	{Ranged{3, 6, Complete}, Ranged{3, 7, Complete}, 4, 1},
	{Ranged{3, 6, Complete}, Ranged{3, 7, Complete}, 5, 1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 6, 1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 7, 1},

	{Ranged{3, 6, Complete}, Ranged{2, 5, Complete}, 2, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 5, Partial5}, 3, -1},
	{Ranged{3, 6, Partial3}, Ranged{3, 5, PartialBoth}, 3, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 5, Complete}, 4, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 5, Partial3}, 5, -1},
	{Ranged{3, 6, Partial5}, Ranged{3, 5, PartialBoth}, 5, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 6, -1},
	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}, 7, -1},
	{Ranged{3, 6, Complete}, Between(3), 3, -3},
	{Ranged{3, 6, Complete}, Between(2), 2, -4},

	{Ambiguous{3, 6}, Ambiguous{3, 6}, 2, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 3, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 4, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 5, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 0},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 0},

	{Ambiguous{3, 6}, Ambiguous{4, 7}, 2, 1},
	{Ambiguous{3, 6}, Ambiguous{4, 7}, 3, 1},
	{Ambiguous{3, 6}, Ambiguous{3, 7}, 4, 1},
	{Ambiguous{3, 6}, Ambiguous{3, 7}, 5, 1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 1},

	{Ambiguous{3, 6}, Ambiguous{2, 5}, 2, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 3, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 4, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 5, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, -1},
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, -1},
	{Ambiguous{3, 6}, Between(3), 3, -4},
	{Ambiguous{3, 6}, Between(2), 2, -4},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 2, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 3, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 4, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 5, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 6, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 7, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 12, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 13, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 14, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 15, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 0},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 0},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 2, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 3, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 7, Complete}, Ranged{14, 17, Complete}}, 4, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 7, Complete}, Ranged{14, 17, Complete}}, 5, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 6, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 7, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 12, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 13, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 17, Complete}}, 14, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 17, Complete}}, 15, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 1},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{2, 5, Complete}, Ranged{12, 15, Complete}}, 2, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 5, Partial5}, Ranged{12, 15, Complete}}, 3, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 5, Complete}, Ranged{12, 15, Complete}}, 4, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 5, Partial3}, Ranged{12, 15, Complete}}, 5, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 6, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 7, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 12, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 15, Partial5}}, 13, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 15, Complete}}, 14, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 15, Partial3}}, 15, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, -1},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, -1},

	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Between(2), Ranged{9, 12, Complete}}, 2, -4},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Joined{Ranged{3, 6, Complete}, Between(12)}, 12, -4},
	{Joined{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Between(2), 2, -14},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 2, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 3, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 4, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 5, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 6, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 7, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 12, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 13, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 14, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 15, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 0},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 0},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 2, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{4, 7, Complete}, Ranged{14, 17, Complete}}, 3, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 7, Complete}, Ranged{14, 17, Complete}}, 4, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 7, Complete}, Ranged{14, 17, Complete}}, 5, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 6, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 7, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 12, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{14, 17, Complete}}, 13, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 17, Complete}}, 14, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 17, Complete}}, 15, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, 1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, 1},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{2, 5, Complete}, Ranged{12, 15, Complete}}, 2, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 5, Partial5}, Ranged{12, 15, Complete}}, 3, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 5, Complete}, Ranged{12, 15, Complete}}, 4, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 5, Partial3}, Ranged{12, 15, Complete}}, 5, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 6, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 7, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{12, 15, Complete}}, 12, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 15, Partial5}}, 13, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 15, Complete}}, 14, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 15, Partial3}}, 15, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 16, -1},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, 17, -1},

	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Between(2), Ranged{9, 12, Complete}}, 2, -4},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Ranged{3, 6, Complete}, Between(12)}, 12, -4},
	{Ordered{Ranged{3, 6, Complete}, Ranged{13, 16, Complete}}, Ordered{Between(2), Between(2)}, 2, -14},
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

var asCompleteTests = []struct {
	in  Location
	out Location
}{
	{Between(3), Between(3)},

	{Ranged{3, 6, Complete}, Ranged{3, 6, Complete}},
	{Ranged{3, 6, Partial5}, Ranged{3, 6, Complete}},
	{Ranged{3, 6, Partial3}, Ranged{3, 6, Complete}},
	{Ranged{3, 6, PartialBoth}, Ranged{3, 6, Complete}},

	{Joined{Ranged{3, 6, Complete}}, Joined{Ranged{3, 6, Complete}}},
	{Joined{Ranged{3, 6, Partial5}}, Joined{Ranged{3, 6, Complete}}},
	{Joined{Ranged{3, 6, Partial3}}, Joined{Ranged{3, 6, Complete}}},
	{Joined{Ranged{3, 6, PartialBoth}}, Joined{Ranged{3, 6, Complete}}},

	{Ordered{Ranged{3, 6, Complete}}, Ordered{Ranged{3, 6, Complete}}},
	{Ordered{Ranged{3, 6, Partial5}}, Ordered{Ranged{3, 6, Complete}}},
	{Ordered{Ranged{3, 6, Partial3}}, Ordered{Ranged{3, 6, Complete}}},
	{Ordered{Ranged{3, 6, PartialBoth}}, Ordered{Ranged{3, 6, Complete}}},
}

func TestAsComplete(t *testing.T) {
	for _, tt := range asCompleteTests {
		out := asComplete(tt.in)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("asComplete(%s) = %s, want %s", locRep(tt.in), locRep(out), locRep(tt.out))
		}
	}
}

var locationReductionTests = []struct {
	in  Location
	out Location
}{
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Range(0, 1), Point(0)},
	{Join(Between(3), Between(3)), Between(3)},
	{Join(Between(3), Point(3)), Point(3)},
	{Join(Between(3), Range(3, 6)), Range(3, 6)},

	{Join(Point(3), Between(4)), Point(3)},
	{Join(Point(3), Point(3)), Point(3)},
	{Join(Point(3), Range(3, 6)), Range(3, 6)},

	{Join(Range(3, 6), Between(6)), Range(3, 6)},
	{Join(Range(3, 6), Point(6)), Range(3, 6)},
	{Join(Range(3, 6), Range(6, 9)), Range(3, 9)},
	{Join(Range(6, 9).Complement(), Range(3, 6).Complement()), Range(3, 9).Complement()},

	{Order(Range(3, 6)), Range(3, 6)},
}

func TestLocationReduction(t *testing.T) {
	for _, tt := range locationReductionTests {
		testutils.Equals(t, tt.in, tt.out)
	}
}

var locationParserPassTests = []struct {
	prs pars.Parser
	loc Location
}{
	{parseBetween, Between(0)},
	{parsePoint, Point(0)},
	{parseRange, Range(0, 2)},
	{parseRange, PartialRange(0, 2, Partial5)},
	{parseRange, PartialRange(0, 2, Partial3)},
	{parseRange, PartialRange(0, 2, PartialBoth)},
	{parseRange, PartialRange(0, 2, Partial3)},
	{parseRange, PartialRange(0, 2, PartialBoth)},
	{parseComplementDefault, Range(0, 2).Complement()},
	{parseJoin, Join(Range(0, 2), Range(3, 5))},
	{parseAmbiguous, Ambiguous{0, 2}},
	{parseOrder, Order(Range(0, 2), Range(2, 4))},
}

var locationParserFailTests = []struct {
	prs pars.Parser
	in  string
}{
	{parseBetween, ""},
	{parseBetween, "?"},
	{parseBetween, "1"},
	{parseBetween, "1?"},
	{parseBetween, "1^?"},
	{parseBetween, "1^3"},

	{parsePoint, ""},
	{parsePoint, "?"},

	{parseRange, ""},
	{parseRange, "?"},
	{parseRange, "1"},
	{parseRange, "1??"},
	{parseRange, "1..?"},

	{parseComplementDefault, ""},
	{parseComplementDefault, "complement?"},
	{parseComplementDefault, "complement(?"},
	{parseComplementDefault, "complement(1..2"},
	{parseComplementDefault, "complement(1..2?"},

	{parseJoin, ""},
	{parseJoin, "join?"},
	{parseJoin, "join("},
	{parseJoin, "join(1..2,?"},
	{parseJoin, "join(1..2,3..5"},
	{parseJoin, "join(1..2,3..5?"},

	{parseOrder, ""},
	{parseOrder, "order?"},
	{parseOrder, "order("},
	{parseOrder, "order(1..2,?"},
	{parseOrder, "order(1..2,3..5"},
	{parseOrder, "order(1..2,3..5?"},

	{parseAmbiguous, ""},
	{parseAmbiguous, "?"},
	{parseAmbiguous, "1"},
	{parseAmbiguous, "1?"},
	{parseAmbiguous, "1.?"},
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
		res, err := parseLocation.Parse(pars.FromString(tt.in))
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
