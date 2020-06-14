package gts

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/go-gts/gts/testutils"
	"github.com/go-pars/pars"
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
	ss  []Span
}{
	{Between(0), "0^1", 0, []Span{{0, 0}}},
	{Point(0), "1", 1, []Span{{0, 1}}},
	{Range(0, 2), "1..2", 2, []Span{{0, 2}}},

	{PartialRange(0, 2, Complete), "1..2", 2, []Span{{0, 2}}},
	{PartialRange(0, 2, Partial5), "<1..2", 2, []Span{{0, 2}}},
	{PartialRange(0, 2, Partial3), "1..>2", 2, []Span{{0, 2}}},
	{PartialRange(0, 2, PartialBoth), "<1..>2", 2, []Span{{0, 2}}},

	{Join(Range(0, 2), Range(3, 5)), "join(1..2,4..5)", 4, []Span{{0, 2}, {3, 5}}},
	{Join(Range(0, 2), Join(Range(3, 5), Range(6, 8))), "join(1..2,4..5,7..8)", 6, []Span{{0, 2}, {3, 5}, {6, 8}}},
	{Join(Point(0), Point(2)), "join(1,3)", 2, []Span{{0, 1}, {2, 1}}},

	{Ambiguous{0, 2}, "1.2", 1, []Span{{0, 1}, {1, 1}}},

	{Order(Range(0, 2), Range(2, 4)), "order(1..2,3..4)", 4, []Span{{0, 2}, {2, 4}}},
	{Order(Range(0, 2), Order(Range(2, 4), Range(4, 6))), "order(1..2,3..4,5..6)", 6, []Span{{0, 2}, {2, 4}, {4, 6}}},
	{Order(Point(0), Point(2)), "order(1,3)", 2, []Span{{0, 1}, {2, 1}}},
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
			testutils.Equals(t, tt.in.Regions(), tt.ss)
			if c, ok := tt.in.(Locatable); ok {
				in := c.Complement()
				if in.Len() != tt.len {
					t.Errorf("%s.Len() = %d, want %d", locRep(in), in.Len(), tt.len)
				}
				testutils.Equals(t, in.Regions(), tt.ss)
			}
		})
	}
}

var locationShiftTests = []struct {
	in, out Location
	i, n    int
	expand  bool
}{
	{Between(0), Between(0), 0, 1, true},
	{Between(0), Between(0), 0, -1, true},
	{Between(1), Between(2), 0, 1, true},
	{Between(1), Between(0), 0, -1, true},

	{Point(0), Point(1), 0, 1, true},
	{Point(0), Between(0), 0, -1, true},
	{Point(0), Point(0), 1, 1, true},
	{Point(0), Point(0), 1, -1, true},
	{Point(1), Point(2), 0, 1, true},
	{Point(1), Point(0), 0, -1, true},

	{Range(0, 2), Range(1, 3), 0, 1, true},
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Range(0, 2), Point(0), 0, -1, true},
	{Range(0, 2), Range(0, 1), 0, -1, true},
	{Range(0, 2), Between(0), 0, -2, true},
	{Range(1, 3), Range(0, 2), 0, -1, true},
	{Range(0, 2), Range(0, 2), 2, 1, true},
	{Range(0, 2), Range(0, 2), 2, -1, true},
	{PartialRange(0, 4, PartialBoth), Join(PartialRange(0, 2, Partial5), PartialRange(3, 5, Partial3)), 2, 1, false},

	{Join(Range(0, 2), Range(3, 5)), Join(Range(1, 3), Range(4, 6)), 0, 1, true},
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Join(Range(0, 2), Range(3, 5)), Join(Point(0), Range(2, 4)), 0, -1},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 1), Range(2, 4)), 0, -1, true},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 2), Range(4, 6)), 2, 1, true},
	{Join(Range(0, 2), Range(3, 5)), Range(0, 4), 2, -1, true},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 2), Range(3, 5)), 5, 1, true},
	{Join(Range(0, 2), Range(3, 5)), Join(Range(0, 2), Range(3, 5)), 5, -1, true},

	{Ambiguous{0, 2}, Ambiguous{1, 3}, 0, 1, true},
	{Ambiguous{0, 2}, Ambiguous{0, 1}, 0, -1, true},
	{Ambiguous{0, 2}, Between(0), 0, -2, true},
	{Ambiguous{1, 3}, Ambiguous{0, 2}, 0, -1, true},
	{Ambiguous{0, 2}, Ambiguous{0, 2}, 2, 1, true},
	{Ambiguous{0, 2}, Ambiguous{0, 2}, 2, -1, true},
	{Ambiguous{0, 4}, Order(Ambiguous{0, 2}, Ambiguous{3, 5}), 2, 1, false},

	{Order(Range(0, 2), Range(3, 5)), Order(Range(1, 3), Range(4, 6)), 0, 1, true},
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Order(Range(0, 2), Range(3, 5)), Order(Point(0), Range(2, 4)), 0, -1},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 1), Range(2, 4)), 0, -1, true},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(4, 6)), 2, 1, true},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(2, 4)), 2, -1, true},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, 1, true},
	{Order(Range(0, 2), Range(3, 5)), Order(Range(0, 2), Range(3, 5)), 5, -1, true},
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
		if !reflect.DeepEqual(tt.in.Shift(tt.i, tt.n, tt.expand), tt.out) {
			t.Errorf(
				"%s.Shift(%d, %d, %t) = %s, want %s",
				locRep(tt.in), tt.i, tt.n, tt.expand,
				locRep(tt.in.Shift(tt.i, tt.n, tt.expand)),
				locRep(tt.out),
			)
		}
		if areLocatable(tt.in, tt.out) {
			if !reflect.DeepEqual(
				tt.in.(Locatable).Complement().Shift(tt.i, tt.n, tt.expand),
				tt.out.(Locatable).Complement(),
			) {
				t.Errorf(
					"%s.Shift(%d, %d, %t) = %s, want %s",
					locRep(tt.in.(Locatable).Complement()), tt.i, tt.n, tt.expand,
					locRep(tt.in.(Locatable).Complement().Shift(tt.i, tt.n, tt.expand)),
					locRep(tt.out.(Locatable).Complement()),
				)
			}
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

func (null NullLocation) Regions() []Span {
	return []Span{{int(null), 0}}
}

func (null NullLocation) Shift(i, n int, expand bool) Location {
	return null
}

func (null NullLocation) Less(loc Location) bool {
	return false
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
	if l, ok := lhs.(Locatable); ok {
		if _, ok := l.(Complemented); !ok {
			locationLessPassTest(t, l.Complement(), rhs)
		}
	}
	if r, ok := rhs.(Locatable); ok {
		if _, ok := r.(Complemented); !ok {
			locationLessPassTest(t, lhs, r.Complement())
		}
	}
}

func locationLessFailTest(t *testing.T, lhs, rhs Location) {
	if lhs.Less(rhs) {
		t.Errorf("expected %s >= %s", locRep(lhs), locRep(rhs))
	}
	if l, ok := lhs.(Locatable); ok {
		if _, ok := l.(Complemented); !ok {
			locationLessFailTest(t, l.Complement(), rhs)
		}
	}
	if r, ok := rhs.(Locatable); ok {
		if _, ok := r.(Complemented); !ok {
			locationLessFailTest(t, lhs, r.Complement())
		}
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

var locatableTests = []struct {
	in  Locatable
	out Sequence
}{
	{Between(0), New(nil, nil, []byte(""))},
	{Point(0), New(nil, nil, []byte("a"))},
	{Range(0, 2), New(nil, nil, []byte("at"))},
	{Join(Range(0, 2), Range(3, 5)), New(nil, nil, []byte("atca"))},
}

func TestLocatable(t *testing.T) {
	seq := New(nil, nil, []byte("atgcatgc"))
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
