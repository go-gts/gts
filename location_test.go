package gts

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-pars/pars"
)

var locationUtilsTests = []struct {
	s0, e0, s1, e1  int
	within, overlap bool
	compare         int
}{
	{3, 6, 3, 6, true, true, 0},    // case 1
	{3, 6, 4, 6, false, true, -1},  // case 2
	{3, 6, 2, 6, true, true, 1},    // case 3
	{3, 6, 3, 7, true, true, -1},   // case 4
	{3, 6, 3, 5, false, true, 1},   // case 5
	{3, 6, 6, 9, false, false, -1}, // case 6
	{3, 6, 0, 3, false, false, 1},  // case 7

	{6, 3, 3, 6, true, true, 0},    // case 8
	{6, 3, 4, 6, false, true, -1},  // case 9
	{6, 3, 2, 6, true, true, 1},    // case 10
	{6, 3, 3, 7, true, true, -1},   // case 11
	{6, 3, 3, 5, false, true, 1},   // case 12
	{6, 3, 6, 9, false, false, -1}, // case 13
	{6, 3, 0, 3, false, false, 1},  // case 14

	{3, 6, 6, 3, true, true, 0},    // case 15
	{3, 6, 6, 4, false, true, -1},  // case 16
	{3, 6, 6, 2, true, true, 1},    // case 17
	{3, 6, 7, 3, true, true, -1},   // case 18
	{3, 6, 5, 3, false, true, 1},   // case 19
	{3, 6, 9, 6, false, false, -1}, // case 20
	{3, 6, 3, 0, false, false, 1},  // case 21
}

func TestLocationUtils(t *testing.T) {
	for i, tt := range locationUtilsTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			isWithin := rangeWithin(tt.s0, tt.e0, tt.s1, tt.e1)
			if isWithin != tt.within {
				t.Errorf(
					"rangeWithin(%d, %d, %d, %d) = %v, expected %v",
					tt.s0, tt.e0, tt.s1, tt.e1, isWithin, tt.within,
				)
			}

			isOverlap := rangeOverlap(tt.s0, tt.e0, tt.s1, tt.e1)
			if isOverlap != tt.overlap {
				t.Errorf(
					"rangeOverlap(%d, %d, %d, %d) = %v, expected %v",
					tt.s0, tt.e0, tt.s1, tt.e1, isOverlap, tt.overlap,
				)
			}

			compare := rangeCompare(tt.s0, tt.e0, tt.s1, tt.e1)
			if compare != tt.compare {
				t.Errorf("rangeCmp(%d, %d, %d, %d) = %d, expected %d",
					tt.s0, tt.e0, tt.s1, tt.e1, compare, tt.compare,
				)
			}
		})
	}
}

var locationSpanTests = []struct {
	in   contiguousLocation
	a, b int
}{
	{Between(3), 3, 3},      // case 1
	{Point(3), 3, 4},        // case 2
	{Range(3, 6), 3, 6},     // case 3
	{Ambiguous{3, 6}, 3, 6}, // case 4
}

func TestLocationSpan(t *testing.T) {
	for i, tt := range locationSpanTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			a, b := tt.in.span()
			if a != tt.a || b != tt.b {
				t.Errorf(
					"%s.span() = (%d, %d), want (%d, %d)",
					tt.in, a, b, tt.a, tt.b,
				)
			}
		})
	}
}

var locationSliceTests = []struct {
	in  locationSlice
	out []Location
}{
	// case 1
	{
		Joined{Range(3, 6), Range(13, 16)},
		[]Location{Range(3, 6), Range(13, 16)},
	},

	// case 2
	{
		Ordered{Range(3, 6), Range(13, 16)},
		[]Location{Range(3, 6), Range(13, 16)},
	},
}

func TestLocationSlice(t *testing.T) {
	for i, tt := range locationSliceTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := tt.in.slice()
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf("%s.slice() = %#v, want %#v", tt.in, out, tt.out)
			}
		})
	}
}

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

var locationLessTests = []struct {
	a, b Location
	out  bool
}{
	{NullLocation(0), NullLocation(0), false}, // case 1
	{NullLocation(0), Range(3, 6), false},     // case 2
	{Range(3, 6), NullLocation(0), true},      // case 3

	{Range(3, 6), Range(2, 5), false}, // case 4
	{Range(3, 6), Range(4, 7), true},  // case 5

	{Range(3, 6), Range(3, 6), false},             // case 6
	{Partial5(3, 6), Partial5(3, 6), false},       // case 7
	{Partial3(3, 6), Partial3(3, 6), false},       // case 8
	{Partial5(3, 6), Partial3(3, 6), false},       // case 9
	{Partial3(3, 6), Partial5(3, 6), false},       // case 10
	{PartialBoth(3, 6), PartialBoth(3, 6), false}, // case 11

	{Partial5(3, 6), Range(3, 6), false},       // case 12
	{Partial3(3, 6), Range(3, 6), false},       // case 13
	{PartialBoth(3, 6), Range(3, 6), false},    // case 14
	{PartialBoth(3, 6), Partial5(3, 6), false}, // case 15
	{PartialBoth(3, 6), Partial3(3, 6), false}, // case 16

	{Range(3, 6), Partial5(3, 6), true},       // case 17
	{Range(3, 6), Partial3(3, 6), true},       // case 18
	{Range(3, 6), PartialBoth(3, 6), true},    // case 19
	{Partial5(3, 6), PartialBoth(3, 6), true}, // case 20
	{Partial3(3, 6), PartialBoth(3, 6), true}, // case 21

	{Range(3, 6), Joined{Range(3, 6), Range(13, 16)}, false}, // case 22
	{Range(3, 6), Joined{Range(4, 7), Range(13, 16)}, true},  // case 23

	{Joined{Range(3, 6), Range(13, 16)}, Range(3, 6), false}, // case 24
	{Joined{Range(3, 6), Range(13, 16)}, Range(4, 7), true},  // case 25
}

func TestLocationLess(t *testing.T) {
	for i, tt := range locationLessTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			a, b := tt.a, tt.b
			out := LocationLess(a, b)
			if out != tt.out {
				t.Errorf(
					"expected %s < %s = %v, want %v",
					locRep(a), locRep(b), out, tt.out,
				)
			}

			a, b = a.Complement(), b.Complement()
			out = LocationLess(a, b)
			if out != tt.out {
				t.Errorf(
					"expected %s < %s = %v, want %v",
					locRep(a), locRep(b), out, tt.out,
				)
			}
		})
	}
}

var locationWithinTests = []struct {
	loc  Location
	l, u int
	out  bool
}{
	{NullLocation(0), 3, 6, false}, // case 1

	{Range(3, 6), 3, 6, true},  // case 2
	{Range(3, 6), 2, 7, true},  // case 3
	{Range(3, 6), 2, 5, false}, // case 4
	{Range(3, 6), 4, 7, false}, // case 5

	{Joined{Range(3, 6), Range(13, 16)}, 3, 6, false}, // case 6
	{Joined{Range(3, 6), Range(13, 16)}, 3, 16, true}, // case 7
}

func TestLocationWithin(t *testing.T) {
	for i, tt := range locationWithinTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			format := "expected %s to be within (%d, %d)"
			if !tt.out {
				format = "expected %s not to be within (%d, %d)"
			}

			loc, l, u := tt.loc, tt.l, tt.u
			out := LocationWithin(loc, l, u)
			if out != tt.out {
				t.Errorf(format, locRep(loc), l, u)
			}

			loc = loc.Complement()
			out = LocationWithin(loc, l, u)
			if out != tt.out {
				t.Errorf(format, locRep(loc), l, u)
			}
		})
	}
}

var locationOverlapTests = []struct {
	loc  Location
	l, u int
	out  bool
}{
	{NullLocation(0), 3, 6, false}, // case 1

	{Range(3, 6), 3, 6, true},  // case 2
	{Range(3, 6), 0, 3, false}, // case 3
	{Range(3, 6), 6, 9, false}, // case 4

	{Joined{Range(3, 6), Range(13, 16)}, 3, 6, true},  // case 5
	{Joined{Range(3, 6), Range(13, 16)}, 6, 9, false}, // case 6
}

func TestLocationOverlap(t *testing.T) {
	for i, tt := range locationOverlapTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			format := "expected %s to overlap (%d, %d)"
			if !tt.out {
				format = "expected %s not to overlap (%d, %d)"
			}

			loc, l, u := tt.loc, tt.l, tt.u
			out := LocationOverlap(loc, l, u)
			if out != tt.out {
				t.Errorf(format, locRep(loc), l, u)
			}

			loc = loc.Complement()
			out = LocationOverlap(loc, l, u)
			if out != tt.out {
				t.Errorf(format, locRep(loc), l, u)
			}
		})
	}
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
		return fmt.Sprintf("Ambiguous(%d, %d)", v.Start, v.End)

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
	{Between(0), "0^1", 0, Segment{0, 0}},   // case 1
	{Point(0), "1", 1, Segment{0, 1}},       // case 2
	{Range(0, 2), "1..2", 2, Segment{0, 2}}, // case 3

	{Range(0, 2), "1..2", 2, Segment{0, 2}},         // case 4
	{Partial5(0, 2), "<1..2", 2, Segment{0, 2}},     // case 5
	{Partial3(0, 2), "1..>2", 2, Segment{0, 2}},     // case 6
	{PartialBoth(0, 2), "<1..>2", 2, Segment{0, 2}}, // case 7

	{Ambiguous{0, 2}, "1.2", 1, Segment{0, 2}}, // case 8

	{Join(Range(0, 2), Range(3, 5)), "join(1..2,4..5)", 4, Regions{Segment{0, 2}, Segment{3, 5}}},                                        // case 9
	{Join(Range(0, 2), Join(Range(3, 5), Range(6, 8))), "join(1..2,4..5,7..8)", 6, Regions{Segment{0, 2}, Segment{3, 5}, Segment{6, 8}}}, // case 10
	{Join(Point(0), Point(2)), "join(1,3)", 2, Regions{Segment{0, 1}, Segment{2, 3}}},                                                    // case 11

	{Order(Range(0, 2), Range(2, 4)), "order(1..2,3..4)", 4, Regions{Segment{0, 2}, Segment{2, 4}}},                                         // case 12
	{Order(Range(0, 2), Order(Range(2, 4), Range(4, 6))), "order(1..2,3..4,5..6)", 6, Regions{Segment{0, 2}, Segment{2, 4}, Segment{4, 6}}}, // case 13
	{Order(Point(0), Point(2)), "order(1,3)", 2, Regions{Segment{0, 1}, Segment{2, 3}}},                                                     // case 14

	{Range(0, 2).Complement(), "complement(1..2)", 2, Segment{2, 0}}, // case 15
}

func TestLocationAccessors(t *testing.T) {
	for i, tt := range locationAccessorTests {
		testutils.RunCase(t, i, func(t *testing.T) {
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

func TestLocationSort(t *testing.T) {
	in := []Location{Range(3, 13), Range(4, 13), Range(6, 14), Range(6, 16)}
	exp := make([]Location, len(in))
	out := make([]Location, len(in))

	copy(exp, in)
	for reflect.DeepEqual(in, exp) {
		rand.Shuffle(len(in), func(i, j int) {
			in[i], in[j] = in[j], in[i]
		})
	}

	copy(out, in)
	sort.Sort(Locations(out))
	if !reflect.DeepEqual(out, exp) {
		t.Errorf("sort.Sort(Locations(%v)) = %v, want %v", in, out, exp)
	}
}

var locationReverseTest = []struct {
	in  Location
	out Location
}{
	{NullLocation(0), NullLocation(0)},                                  // case 1
	{Between(0), Between(9)},                                            // case 2
	{Point(0), Point(9)},                                                // case 3
	{Range(0, 3), Range(7, 10)},                                         // case 4
	{Partial5(0, 3), Partial3(7, 10)},                                   // case 5
	{Partial3(0, 3), Partial5(7, 10)},                                   // case 6
	{PartialBoth(0, 3), PartialBoth(7, 10)},                             // case 7
	{Join(Range(0, 3), Range(5, 8)), Join(Range(2, 5), Range(7, 10))},   // case 8
	{Range(0, 3).Complement(), Range(7, 10).Complement()},               // case 9
	{Ambiguous{0, 3}, Ambiguous{7, 10}},                                 // case 10
	{Order(Range(0, 3), Range(5, 8)), Order(Range(2, 5), Range(7, 10))}, // case 11
}

func TestLocationReverse(t *testing.T) {
	for i, tt := range locationReverseTest {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := tt.in.Reverse(10)
			testutils.Equals(t, out, tt.out)
		})
	}
}

var locationNormalizeTest = []struct {
	in  Location
	out Location
}{
	{NullLocation(0), NullLocation(0)},                                   // case 1
	{Between(10), Between(0)},                                            // case 2
	{Point(10), Point(0)},                                                // case 3
	{Range(10, 13), Range(0, 3)},                                         // case 4
	{Range(3, 13), Range(0, 10)},                                         // case 5
	{Range(8, 12), Join(Range(8, 10), Range(0, 2))},                      // case 6
	{PartialBoth(8, 12), Join(Partial5(8, 10), Partial3(0, 2))},          // case 7
	{Join(Range(10, 13), Range(5, 8)), Join(Range(0, 3), Range(5, 8))},   // case 8
	{Range(10, 13).Complement(), Range(0, 3).Complement()},               // case 9
	{Ambiguous{10, 13}, Ambiguous{0, 3}},                                 // case 10
	{Order(Range(10, 13), Range(5, 8)), Order(Range(0, 3), Range(5, 8))}, // case 11
}

func TestLocationNormalize(t *testing.T) {
	for i, tt := range locationNormalizeTest {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := tt.in.Normalize(10)
			testutils.Equals(t, out, tt.out)
		})
	}
}

var locationShiftTests = []struct {
	in, out Location
	i, n    int
}{
	{Between(3), Between(3), 2, 0},  // case 1
	{Between(3), Between(3), 3, 0},  // case 2
	{Between(3), Between(3), 4, 0},  // case 3
	{Between(3), Between(4), 2, 1},  // case 4
	{Between(3), Between(3), 3, 1},  // case 5
	{Between(3), Between(3), 4, 1},  // case 6
	{Between(3), Between(2), 2, -1}, // case 7
	{Between(3), Between(3), 3, -1}, // case 8
	{Between(3), Between(3), 4, -1}, // case 9

	{Point(3), Point(3), 2, 0},    // case 10
	{Point(3), Point(3), 3, 0},    // case 11
	{Point(3), Point(3), 4, 0},    // case 12
	{Point(3), Point(4), 2, 1},    // case 13
	{Point(3), Point(4), 3, 1},    // case 14
	{Point(3), Point(3), 4, 1},    // case 15
	{Point(3), Point(2), 2, -1},   // case 16
	{Point(3), Between(3), 3, -1}, // case 17
	{Point(3), Point(3), 4, -1},   // case 18

	{Range(3, 6), Range(3, 6), 2, 0}, // case 19
	{Range(3, 6), Range(3, 6), 3, 0}, // case 20
	{Range(3, 6), Range(3, 6), 4, 0}, // case 21
	{Range(3, 6), Range(3, 6), 5, 0}, // case 22
	{Range(3, 6), Range(3, 6), 6, 0}, // case 23
	{Range(3, 6), Range(3, 6), 7, 0}, // case 24

	{Range(3, 6), Range(4, 7), 2, 1},                                // case 25
	{Range(3, 6), Range(4, 7), 3, 1},                                // case 26
	{Range(3, 6), Join(Range(3, 4), Range(5, 7)), 4, 1},             // case 27
	{Partial5(3, 6), Join(Partial5(3, 4), Range(5, 7)), 4, 1},       // case 28
	{Partial3(3, 6), Join(Range(3, 4), Partial3(5, 7)), 4, 1},       // case 29
	{PartialBoth(3, 6), Join(Partial5(3, 4), Partial3(5, 7)), 4, 1}, // case 30
	{Range(3, 6), Join(Range(3, 5), Range(6, 7)), 5, 1},             // case 31
	{Partial5(3, 6), Join(Partial5(3, 5), Range(6, 7)), 5, 1},       // case 32
	{Partial3(3, 6), Join(Range(3, 5), Partial3(6, 7)), 5, 1},       // case 33
	{PartialBoth(3, 6), Join(Partial5(3, 5), Partial3(6, 7)), 5, 1}, // case 34
	{Range(3, 6), Range(3, 6), 6, 1},                                // case 35
	{Range(3, 6), Range(3, 6), 7, 1},                                // case 36

	{Range(3, 6), Range(2, 5), 2, -1},          // case 37
	{Range(3, 6), Partial5(3, 5), 3, -1},       // case 38
	{Partial3(3, 6), PartialBoth(3, 5), 3, -1}, // case 39
	{Range(3, 6), Range(3, 5), 4, -1},          // case 40
	{Range(3, 6), Partial3(3, 5), 5, -1},       // case 41
	{Partial5(3, 6), PartialBoth(3, 5), 5, -1}, // case 42
	{Range(3, 6), Range(3, 6), 6, -1},          // case 43
	{Range(3, 6), Range(3, 6), 7, -1},          // case 44
	{Range(3, 6), Between(2), 2, -4},           // case 45

	{Ambiguous{3, 6}, Ambiguous{3, 6}, 2, 0}, // case 46
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 3, 0}, // case 47
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 4, 0}, // case 48
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 5, 0}, // case 49
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 0}, // case 50
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 0}, // case 51

	{Ambiguous{3, 6}, Ambiguous{4, 7}, 2, 1},                         // case 52
	{Ambiguous{3, 6}, Ambiguous{4, 7}, 3, 1},                         // case 53
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1}, // case 54
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1}, // case 55
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1}, // case 56
	{Ambiguous{3, 6}, Order(Ambiguous{3, 4}, Ambiguous{5, 7}), 4, 1}, // case 57
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1}, // case 58
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1}, // case 59
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1}, // case 60
	{Ambiguous{3, 6}, Order(Ambiguous{3, 5}, Ambiguous{6, 7}), 5, 1}, // case 61
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 1},                         // case 62
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 1},                         // case 63

	{Ambiguous{3, 6}, Ambiguous{2, 5}, 2, -1}, // case 64
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 3, -1}, // case 65
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 3, -1}, // case 66
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 4, -1}, // case 67
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 5, -1}, // case 68
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 5, -1}, // case 69
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, -1}, // case 70
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, -1}, // case 71
	{Ambiguous{3, 6}, Between(2), 2, -4},      // case 72

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 2, 0},  // case 73
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 3, 0},  // case 74
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 4, 0},  // case 75
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 5, 0},  // case 76
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 6, 0},  // case 77
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 7, 0},  // case 78
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 12, 0}, // case 79
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 13, 0}, // case 80
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 14, 0}, // case 81
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 15, 0}, // case 82
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 16, 0}, // case 83
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 17, 0}, // case 84

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(4, 7), Range(14, 17)}, 2, 1},                 // case 85
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(4, 7), Range(14, 17)}, 3, 1},                 // case 86
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 4), Range(5, 7), Range(14, 17)}, 4, 1},    // case 87
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 5), Range(6, 7), Range(14, 17)}, 5, 1},    // case 88
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 6, 1},                 // case 89
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 7, 1},                 // case 90
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 12, 1},                // case 91
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 13, 1},                // case 92
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 14), Range(15, 17)}, 14, 1}, // case 93
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 15), Range(16, 17)}, 15, 1}, // case 94
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 16, 1},                // case 95
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 17, 1},                // case 96

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(2, 5), Range(12, 15)}, 2, -1},     // case 97
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Partial5(3, 5), Range(12, 15)}, 3, -1},  // case 98
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 5), Range(12, 15)}, 4, -1},     // case 99
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Partial3(3, 5), Range(12, 15)}, 5, -1},  // case 100
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(12, 15)}, 6, -1},     // case 101
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(12, 15)}, 7, -1},     // case 102
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(12, 15)}, 12, -1},    // case 103
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Partial5(13, 15)}, 13, -1}, // case 104
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 15)}, 14, -1},    // case 105
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Partial3(13, 15)}, 15, -1}, // case 106
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 16, -1},    // case 107
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 17, -1},    // case 108

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Between(2), Range(9, 12)}, 2, -4},  // case 109
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Between(12)}, 12, -4}, // case 110
	{Joined{Range(3, 6), Range(13, 16)}, Between(2), 2, -14},                       // case 111

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 2, 0},  // case 112
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 3, 0},  // case 113
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 4, 0},  // case 114
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 5, 0},  // case 115
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 6, 0},  // case 116
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 7, 0},  // case 117
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 12, 0}, // case 118
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 13, 0}, // case 119
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 14, 0}, // case 120
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 15, 0}, // case 121
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 16, 0}, // case 122
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 17, 0}, // case 123

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(4, 7), Range(14, 17)}, 2, 1},                         // case 124
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(4, 7), Range(14, 17)}, 3, 1},                         // case 125
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Joined{Range(3, 4), Range(5, 7)}, Range(14, 17)}, 4, 1},    // case 126
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Joined{Range(3, 5), Range(6, 7)}, Range(14, 17)}, 5, 1},    // case 127
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 6, 1},                         // case 128
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 7, 1},                         // case 129
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 12, 1},                        // case 130
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 13, 1},                        // case 131
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Joined{Range(13, 14), Range(15, 17)}}, 14, 1}, // case 132
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Joined{Range(13, 15), Range(16, 17)}}, 15, 1}, // case 133
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 16, 1},                        // case 134
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 17, 1},                        // case 135

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(2, 5), Range(12, 15)}, 2, -1},     // case 136
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Partial5(3, 5), Range(12, 15)}, 3, -1},  // case 137
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 5), Range(12, 15)}, 4, -1},     // case 138
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Partial3(3, 5), Range(12, 15)}, 5, -1},  // case 139
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(12, 15)}, 6, -1},     // case 140
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(12, 15)}, 7, -1},     // case 141
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(12, 15)}, 12, -1},    // case 142
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Partial5(13, 15)}, 13, -1}, // case 143
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 15)}, 14, -1},    // case 144
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Partial3(13, 15)}, 15, -1}, // case 145
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 16, -1},    // case 146
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 17, -1},    // case 147

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Between(2), Range(9, 12)}, 2, -4},  // case 148
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Between(12)}, 12, -4}, // case 149
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Between(2), Between(2)}, 2, -14},   // case 150
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
	{Between(3), Between(3), 2, 0},  // case 1
	{Between(3), Between(3), 3, 0},  // case 2
	{Between(3), Between(3), 4, 0},  // case 3
	{Between(3), Between(4), 2, 1},  // case 4
	{Between(3), Between(3), 3, 1},  // case 5
	{Between(3), Between(3), 4, 1},  // case 6
	{Between(3), Between(2), 2, -1}, // case 7
	{Between(3), Between(3), 3, -1}, // case 8
	{Between(3), Between(3), 4, -1}, // case 9

	{Point(3), Point(3), 2, 0},    // case 10
	{Point(3), Point(3), 3, 0},    // case 11
	{Point(3), Point(3), 4, 0},    // case 12
	{Point(3), Point(4), 2, 1},    // case 13
	{Point(3), Point(4), 3, 1},    // case 14
	{Point(3), Point(3), 4, 1},    // case 15
	{Point(3), Point(2), 2, -1},   // case 16
	{Point(3), Between(3), 3, -1}, // case 17
	{Point(3), Point(3), 4, -1},   // case 18

	{Range(3, 6), Range(3, 6), 2, 0}, // case 19
	{Range(3, 6), Range(3, 6), 3, 0}, // case 20
	{Range(3, 6), Range(3, 6), 4, 0}, // case 21
	{Range(3, 6), Range(3, 6), 5, 0}, // case 22
	{Range(3, 6), Range(3, 6), 6, 0}, // case 23
	{Range(3, 6), Range(3, 6), 7, 0}, // case 24

	{Range(3, 6), Range(4, 7), 2, 1}, // case 25
	{Range(3, 6), Range(4, 7), 3, 1}, // case 26
	{Range(3, 6), Range(3, 7), 4, 1}, // case 27
	{Range(3, 6), Range(3, 7), 5, 1}, // case 28
	{Range(3, 6), Range(3, 6), 6, 1}, // case 29
	{Range(3, 6), Range(3, 6), 7, 1}, // case 30

	{Range(3, 6), Range(2, 5), 2, -1},          // case 31
	{Range(3, 6), Partial5(3, 5), 3, -1},       // case 32
	{Partial3(3, 6), PartialBoth(3, 5), 3, -1}, // case 33
	{Range(3, 6), Range(3, 5), 4, -1},          // case 34
	{Range(3, 6), Partial3(3, 5), 5, -1},       // case 35
	{Partial5(3, 6), PartialBoth(3, 5), 5, -1}, // case 36
	{Range(3, 6), Range(3, 6), 6, -1},          // case 37
	{Range(3, 6), Range(3, 6), 7, -1},          // case 38
	{Range(3, 6), Between(3), 3, -3},           // case 39
	{Range(3, 6), Between(2), 2, -4},           // case 40

	{Ambiguous{3, 6}, Ambiguous{3, 6}, 2, 0}, // case 41
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 3, 0}, // case 42
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 4, 0}, // case 43
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 5, 0}, // case 44
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 0}, // case 45
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 0}, // case 46

	{Ambiguous{3, 6}, Ambiguous{4, 7}, 2, 1}, // case 47
	{Ambiguous{3, 6}, Ambiguous{4, 7}, 3, 1}, // case 48
	{Ambiguous{3, 6}, Ambiguous{3, 7}, 4, 1}, // case 49
	{Ambiguous{3, 6}, Ambiguous{3, 7}, 5, 1}, // case 50
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, 1}, // case 51
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, 1}, // case 52

	{Ambiguous{3, 6}, Ambiguous{2, 5}, 2, -1}, // case 53
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 3, -1}, // case 54
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 4, -1}, // case 55
	{Ambiguous{3, 6}, Ambiguous{3, 5}, 5, -1}, // case 56
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 6, -1}, // case 57
	{Ambiguous{3, 6}, Ambiguous{3, 6}, 7, -1}, // case 58
	{Ambiguous{3, 6}, Between(3), 3, -4},      // case 59
	{Ambiguous{3, 6}, Between(2), 2, -4},      // case 60

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 2, 0},  // case 61
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 3, 0},  // case 62
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 4, 0},  // case 63
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 5, 0},  // case 64
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 6, 0},  // case 65
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 7, 0},  // case 66
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 12, 0}, // case 67
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 13, 0}, // case 68
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 14, 0}, // case 69
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 15, 0}, // case 70
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 16, 0}, // case 71
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 17, 0}, // case 72

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(4, 7), Range(14, 17)}, 2, 1},  // case 73
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(4, 7), Range(14, 17)}, 3, 1},  // case 74
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 7), Range(14, 17)}, 4, 1},  // case 75
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 7), Range(14, 17)}, 5, 1},  // case 76
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 6, 1},  // case 77
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 7, 1},  // case 78
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 12, 1}, // case 79
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(14, 17)}, 13, 1}, // case 80
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 17)}, 14, 1}, // case 81
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 17)}, 15, 1}, // case 82
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 16, 1}, // case 83
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 17, 1}, // case 84

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(2, 5), Range(12, 15)}, 2, -1},     // case 85
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Partial5(3, 5), Range(12, 15)}, 3, -1},  // case 86
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 5), Range(12, 15)}, 4, -1},     // case 87
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Partial3(3, 5), Range(12, 15)}, 5, -1},  // case 88
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(12, 15)}, 6, -1},     // case 89
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(12, 15)}, 7, -1},     // case 90
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(12, 15)}, 12, -1},    // case 91
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Partial5(13, 15)}, 13, -1}, // case 92
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 15)}, 14, -1},    // case 93
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Partial3(13, 15)}, 15, -1}, // case 94
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 16, -1},    // case 95
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Range(13, 16)}, 17, -1},    // case 96

	{Joined{Range(3, 6), Range(13, 16)}, Joined{Between(2), Range(9, 12)}, 2, -4},  // case 97
	{Joined{Range(3, 6), Range(13, 16)}, Joined{Range(3, 6), Between(12)}, 12, -4}, // case 98
	{Joined{Range(3, 6), Range(13, 16)}, Between(2), 2, -14},                       // case 99

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 2, 0},  // case 100
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 3, 0},  // case 101
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 4, 0},  // case 102
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 5, 0},  // case 103
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 6, 0},  // case 104
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 7, 0},  // case 105
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 12, 0}, // case 106
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 13, 0}, // case 107
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 14, 0}, // case 108
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 15, 0}, // case 109
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 16, 0}, // case 110
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 17, 0}, // case 111

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(4, 7), Range(14, 17)}, 2, 1},  // case 112
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(4, 7), Range(14, 17)}, 3, 1},  // case 113
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 7), Range(14, 17)}, 4, 1},  // case 114
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 7), Range(14, 17)}, 5, 1},  // case 115
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 6, 1},  // case 116
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 7, 1},  // case 117
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 12, 1}, // case 118
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(14, 17)}, 13, 1}, // case 119
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 17)}, 14, 1}, // case 120
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 17)}, 15, 1}, // case 121
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 16, 1}, // case 122
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 17, 1}, // case 123

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(2, 5), Range(12, 15)}, 2, -1},     // case 124
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Partial5(3, 5), Range(12, 15)}, 3, -1},  // case 125
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 5), Range(12, 15)}, 4, -1},     // case 126
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Partial3(3, 5), Range(12, 15)}, 5, -1},  // case 127
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(12, 15)}, 6, -1},     // case 128
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(12, 15)}, 7, -1},     // case 129
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(12, 15)}, 12, -1},    // case 130
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Partial5(13, 15)}, 13, -1}, // case 131
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 15)}, 14, -1},    // case 132
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Partial3(13, 15)}, 15, -1}, // case 133
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 16, -1},    // case 134
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Range(13, 16)}, 17, -1},    // case 135

	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Between(2), Range(9, 12)}, 2, -4},  // case 136
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Range(3, 6), Between(12)}, 12, -4}, // case 137
	{Ordered{Range(3, 6), Range(13, 16)}, Ordered{Between(2), Between(2)}, 2, -14},   // case 138
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
	{Between(3), Between(3)}, // case 1

	{Range(3, 6), Range(3, 6)},       // case 2
	{Partial5(3, 6), Range(3, 6)},    // case 3
	{Partial3(3, 6), Range(3, 6)},    // case 4
	{PartialBoth(3, 6), Range(3, 6)}, // case 5

	{Joined{Range(3, 6)}, Joined{Range(3, 6)}},       // case 6
	{Joined{Partial5(3, 6)}, Joined{Range(3, 6)}},    // case 7
	{Joined{Partial3(3, 6)}, Joined{Range(3, 6)}},    // case 8
	{Joined{PartialBoth(3, 6)}, Joined{Range(3, 6)}}, // case 9

	{Ordered{Range(3, 6)}, Ordered{Range(3, 6)}},       // case 10
	{Ordered{Partial5(3, 6)}, Ordered{Range(3, 6)}},    // case 11
	{Ordered{Partial3(3, 6)}, Ordered{Range(3, 6)}},    // case 12
	{Ordered{PartialBoth(3, 6)}, Ordered{Range(3, 6)}}, // case 13
}

func TestAsComplete(t *testing.T) {
	for i, tt := range asCompleteTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := AsComplete(tt.in)
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf("asComplete(%s) = %s, want %s", locRep(tt.in), locRep(out), locRep(tt.out))
			}
		})
	}
}

var locationReductionTests = []struct {
	in  Location
	out Location
}{
	// DISCUSS: should a complete, one base range be reduced to a Point?
	// {Range(0, 1), Point(0)},
	{Join(Between(3), Between(3)), Between(3)},   // case 1
	{Join(Between(3), Point(3)), Point(3)},       // case 2
	{Join(Between(3), Range(3, 6)), Range(3, 6)}, // case 3

	{Join(Point(3), Between(4)), Point(3)},     // case 4
	{Join(Point(3), Point(3)), Point(3)},       // case 5
	{Join(Point(3), Range(3, 6)), Range(3, 6)}, // case 6

	{Join(Range(3, 6), Between(6)), Range(3, 6)},                                         // case 7
	{Join(Range(3, 6), Point(6)), Range(3, 6)},                                           // case 8
	{Join(Range(3, 6), Range(6, 9)), Range(3, 9)},                                        // case 9
	{Join(Range(6, 9).Complement(), Range(3, 6).Complement()), Range(3, 9).Complement()}, // case 10

	{Order(Range(3, 6)), Range(3, 6)}, // case 11
}

func TestLocationReduction(t *testing.T) {
	for i, tt := range locationReductionTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			testutils.Equals(t, tt.in, tt.out)
		})
	}
}

var locationStrandTests = []struct {
	in  Location
	out Strand
}{
	{Range(3, 6), StrandForward},                                                     // case 1
	{Complemented{Range(3, 6)}, StrandReverse},                                       // case 2
	{Joined{Range(3, 6), Range(13, 16)}, StrandForward},                              // case 3
	{Ordered{Range(3, 6), Range(13, 16)}, StrandForward},                             // case 4
	{Joined{Complemented{Range(3, 6)}, Complemented{Range(13, 16)}}, StrandReverse},  // case 5
	{Ordered{Complemented{Range(3, 6)}, Complemented{Range(13, 16)}}, StrandReverse}, // case 6
	{Joined{Complemented{Range(3, 6)}, Range(13, 16)}, StrandBoth},                   // case 7
	{Ordered{Complemented{Range(3, 6)}, Range(13, 16)}, StrandBoth},                  // case 8
	{Joined{Range(3, 6), Complemented{Range(13, 16)}}, StrandBoth},                   // case 9
	{Ordered{Range(3, 6), Complemented{Range(13, 16)}}, StrandBoth},                  // case 10
	{Joined{Joined{Range(3, 6), Complemented{Range(13, 16)}}}, StrandBoth},           // case 11
}

func TestLocationStrand(t *testing.T) {
	for i, tt := range locationStrandTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := CheckStrand(tt.in)
			testutils.Equals(t, out, tt.out)
		})
	}
}

var locationParserPassTests = []struct {
	prs pars.Parser
	loc Location
}{
	{parseBetween, Between(0)},                         // case 1
	{parsePoint, Point(0)},                             // case 2
	{parseRange, Range(0, 2)},                          // case 3
	{parseRange, Partial5(0, 2)},                       // case 4
	{parseRange, Partial3(0, 2)},                       // case 5
	{parseRange, PartialBoth(0, 2)},                    // case 6
	{parseRange, Partial3(0, 2)},                       // case 7
	{parseRange, PartialBoth(0, 2)},                    // case 8
	{parseComplementDefault, Range(0, 2).Complement()}, // case 9
	{parseJoin, Join(Range(0, 2), Range(3, 5))},        // case 10
	{parseAmbiguous, Ambiguous{0, 2}},                  // case 11
	{parseOrder, Order(Range(0, 2), Range(2, 4))},      // case 12
}

var locationParserFailTests = []struct {
	prs pars.Parser
	in  string
}{
	{parseBetween, ""},    // case 0
	{parseBetween, "?"},   // case 1
	{parseBetween, "1"},   // case 2
	{parseBetween, "1?"},  // case 3
	{parseBetween, "1^?"}, // case 4
	{parseBetween, "1^3"}, // case 5

	{parsePoint, ""},  // case 6
	{parsePoint, "?"}, // case 7

	{parseRange, ""},     // case 8
	{parseRange, "?"},    // case 9
	{parseRange, "1"},    // case 10
	{parseRange, "1??"},  // case 11
	{parseRange, "1.."},  // case 12
	{parseRange, "1..?"}, // case 13

	{parseComplementDefault, ""},                 // case 14
	{parseComplementDefault, "complement?"},      // case 15
	{parseComplementDefault, "complement(?"},     // case 16
	{parseComplementDefault, "complement(1..2"},  // case 17
	{parseComplementDefault, "complement(1..2?"}, // case 18

	{parseJoin, ""},                // case 19
	{parseJoin, "join?"},           // case 20
	{parseJoin, "join("},           // case 21
	{parseJoin, "join(1..2,?"},     // case 22
	{parseJoin, "join(1..2,3..5"},  // case 23
	{parseJoin, "join(1..2,3..5?"}, // case 24

	{parseOrder, ""},                 // case 25
	{parseOrder, "order?"},           // case 26
	{parseOrder, "order("},           // case 27
	{parseOrder, "order(1..2,?"},     // case 28
	{parseOrder, "order(1..2,3..5"},  // case 29
	{parseOrder, "order(1..2,3..5?"}, // case 30

	{parseAmbiguous, ""},    // case 31
	{parseAmbiguous, "?"},   // case 32
	{parseAmbiguous, "1"},   // case 33
	{parseAmbiguous, "1?"},  // case 34
	{parseAmbiguous, "1.?"}, // case 35
}

func TestLocationParsers(t *testing.T) {
	for i, tt := range locationParserPassTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			prs := pars.Exact(tt.prs)
			in := tt.loc.String()
			res, err := prs.Parse(pars.FromString(in))
			if err != nil {
				t.Errorf("while parsing %q got: %v", in, err)
				return
			}
			out, ok := res.Value.(Location)
			if !ok {
				t.Errorf("parsed result is of type `%T`, want Location", res.Value)
				return
			}
			if !reflect.DeepEqual(out, tt.loc) {
				t.Errorf("parser output is %s, want %s", locRep(out), locRep(tt.loc))
			}
		})
	}

	for i, tt := range locationParserFailTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			prs := pars.Exact(tt.prs)
			_, err := prs.Parse(pars.FromString(tt.in))
			if err == nil {
				t.Errorf("expected error while parsing %q", tt.in)
			}
		})
	}
}

var locationParserTests = []struct {
	in  string
	out Location
}{
	{"0^1", Between(0)},                                    // case 1
	{"1", Point(0)},                                        // case 2
	{"1..2", Range(0, 2)},                                  // case 3
	{"<1..2", Partial5(0, 2)},                              // case 4
	{"1..>2", Partial3(0, 2)},                              // case 5
	{"<1..>2", PartialBoth(0, 2)},                          // case 6
	{"1..2>", Partial3(0, 2)},                              // case 7
	{"<1..2>", PartialBoth(0, 2)},                          // case 8
	{"join(1..2,4..5)", Join(Range(0, 2), Range(3, 5))},    // case 9
	{"1.2", Ambiguous{0, 2}},                               // case 10
	{"order(1..2,3..4)", Order(Range(0, 2), Range(2, 4))},  // case 11
	{"order(1..2, 3..4)", Order(Range(0, 2), Range(2, 4))}, // case 12
}

func TestLocationParser(t *testing.T) {
	for i, tt := range locationParserTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			res, err := ParseLocation.Parse(pars.FromString(tt.in))
			if err != nil {
				t.Errorf("failed to parse %q: %v", tt.in, err)
				return
			}
			out, ok := res.Value.(Location)
			if !ok {
				t.Errorf("parsed result is of type `%T`, want Location", res.Value)
				return
			}
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf("parsed %q: expected %s, got %s", tt.in, locRep(tt.out), locRep(out))
			}
		})
	}
}

func TestAsLocation(t *testing.T) {
	for i, tt := range locationParserTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out, err := AsLocation(tt.in)
			if err != nil {
				t.Errorf("while parsing %q got: %v", tt.in, err)
				return
			}
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf("parser output is %s, want %s", locRep(out), locRep(tt.out))
			}
		})
	}
	if _, err := AsLocation(""); err == nil {
		t.Errorf("expected error in AsLocation(%q)", "")
	}
}

func TestLocationPanics(t *testing.T) {
	testutils.Panics(t, func() { Range(2, 0) })
	testutils.Panics(t, func() { Join() })
	testutils.Panics(t, func() { Order() })
}
