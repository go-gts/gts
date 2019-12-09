package gts

import (
	"testing"
)

func ints(is ...int) []int { return is }

func TestLocationLess(t *testing.T) {
	values := []string{
		"1",
		"1..42",
		"42",
		"42..723",
		"723",
	}

	for i := range values {
		for j := range values {
			a, err := AsLocation(values[i])
			if err != nil {
				t.Errorf("AsLocation(%q): %v", values[i], err)
				return
			}
			b, err := AsLocation(values[j])
			if err != nil {
				t.Errorf("AsLocation(%q): %v", values[j], err)
				return
			}
			if LocationLess(a, b) != (i < j) {
				t.Errorf("%s >= %s, want %s < %s", a, b, a, b)
			}
			if LocationLess(b, a) != (j < i) {
				t.Errorf("%s <= %s, want %s > %s", a, b, a, b)
			}
		}
	}
}

var locateTests = []struct {
	in  string
	out Sequence
}{
	{"1", Seq("a")},
	{"2", Seq("t")},
	{"3", Seq("g")},
	{"4", Seq("c")},

	{"1..4", Seq("atgc")},
	{"2..5", Seq("tgca")},
	{"3..6", Seq("gcat")},

	{"1.4", Seq("atgc")},
	{"2.5", Seq("tgca")},
	{"3.6", Seq("gcat")},

	{"1^4", Seq("atgc")},
	{"2^5", Seq("tgca")},
	{"3^6", Seq("gcat")},

	{"complement(1..4)", Complement(Seq("atgc"))},
	{"complement(2..5)", Complement(Seq("tgca"))},
	{"complement(3..6)", Complement(Seq("gcat"))},

	{"join(1..2,4..5,7..8)", Seq("atcagc")},
	{"join(1..3,6..8)", Seq("atgtgc")},

	{"order(1..2,4..5,7..8)", Seq("atcagc")},
	{"order(1..3,6..8)", Seq("atgtgc")},
}

func TestLocate(t *testing.T) {
	in := Seq("atgcatgc")
	for _, tt := range locateTests {
		loc, err := AsLocation(tt.in)
		if err != nil {
			t.Errorf("AsLocation(%q): %v", tt.in, err)
			return
		}
		out := loc.Locate(in)
		if !same(out, tt.out) {
			t.Errorf("loc.Locate(%q) = %q, want %q", in, out, tt.out)
		}
	}
}

var mapTests = []struct {
	in  string
	out []int
}{
	{"1", ints(0)},
	{"2", ints(1)},
	{"3", ints(2)},
	{"4", ints(3)},

	{"1..4", ints(0, 1, 2, 3)},
	{"3..6", ints(2, 3, 4, 5)},
	{"5..8", ints(4, 5, 6, 7)},

	{"1.4", ints(0, 1, 2, 3)},
	{"3.6", ints(2, 3, 4, 5)},
	{"5.8", ints(4, 5, 6, 7)},

	{"1^4", ints(0, 1, 2, 3)},
	{"3^6", ints(2, 3, 4, 5)},
	{"5^8", ints(4, 5, 6, 7)},

	{"complement(1..4)", ints(0, 1, 2, 3)},
	{"complement(3..6)", ints(2, 3, 4, 5)},
	{"complement(5..8)", ints(4, 5, 6, 7)},

	{"join(1..2,4..5,7..8)", ints(0, 1, 3, 4, 6, 7)},
	{"join(1..3,6..8)", ints(0, 1, 2, 5, 6, 7)},

	{"order(1..2,4..5,7..8)", ints(0, 1, 3, 4, 6, 7)},
	{"order(1..3,6..8)", ints(0, 1, 2, 5, 6, 7)},
}

func TestMap(t *testing.T) {
	for _, tt := range mapTests {
		loc, err := AsLocation(tt.in)
		if err != nil {
			t.Errorf("AsLocation(%q): %v", tt.in, err)
			continue
		}
		if loc.Len() != len(tt.out) {
			t.Errorf("loc.Len() = %d, want %d", loc.Len(), len(tt.out))
		}
		for i := range tt.out {
			if loc.Map(i) != tt.out[i] {
				t.Errorf("loc.Map(%d) = %d, want %d", i, loc.Map(i), tt.out[i])
			}
		}
		PanicTest(t, func(t *testing.T) {
			t.Helper()
			loc.Map(-1)
		})
		PanicTest(t, func(t *testing.T) {
			t.Helper()
			loc.Map(loc.Len() + 1)
		})
	}
}

var shiftTests = []struct {
	in, up, down string
	ok           bool
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

func TestLocationsShift(t *testing.T) {
	for _, tt := range shiftTests {
		loc, err := AsLocation(tt.in)
		if err != nil {
			t.Errorf("AsLocation(%q): %v", tt.in, err)
			return
		}
		if !loc.Shift(1, 0) {
			t.Error("loc.Shift(1, 0) = false, want true")
		}

		up, err := AsLocation(tt.up)
		if err != nil {
			t.Errorf("AsLocation(%q): %v", tt.up, err)
			return
		}
		if !loc.Shift(1, 1) {
			t.Error("loc.Shift(1, 1) = false, want true")
		}
		if !same(loc, up) {
			t.Errorf("%#v != %#v", loc, up)
		}

		if !loc.Shift(1, -1) {
			t.Error("loc.Shift(1, -1) = false, want true")
		}

		down, err := AsLocation(tt.down)
		if err != nil {
			t.Errorf("AsLocation(%q): %v", tt.down, err)
			return
		}
		if ok := loc.Shift(1, -1); ok != tt.ok {
			t.Errorf("loc.Shift(1, -1) = %t, want %t", ok, tt.ok)
		}
		if !same(loc, down) {
			t.Errorf("%#v != %#v", loc, down)
		}
	}
}

var locationTests = []string{
	// point
	"42",

	// range
	"1..42",
	"<1..42",
	"1..>42",
	"<1..>42",

	// ambiguous
	"1.42",

	// between
	"1^42",

	// complement
	"complement(1..42)",
	"complement(<1..42)",
	"complement(1..>42)",
	"complement(<1..>42)",
	"complement(join(1..42,346..723))",
	"complement(join(<1..42,346..723))",
	"complement(join(1..>42,346..723))",
	"complement(join(<1..>42,346..723))",
	"complement(join(1..42,<346..723))",
	"complement(join(1..42,346..>723))",
	"complement(join(1..42,<346..>723))",
	"complement(join(<1..42,<346..723))",
	"complement(join(1..>42,346..>723))",
	"complement(join(<1..>42,<346..>723))",
	"complement(join(1..42,complement(346..723)))",
	"complement(join(complement(1..42),346..723))",
	"complement(join(complement(1..42),complement(346..723)))",

	// join
	"join(1..42,346..723)",
	"join(<1..42,346..723)",
	"join(1..>42,346..723)",
	"join(<1..>42,346..723)",
	"join(1..42,<346..723)",
	"join(1..42,346..>723)",
	"join(1..42,<346..>723)",
	"join(<1..42,<346..723)",
	"join(1..>42,346..>723)",
	"join(<1..>42,<346..>723)",
	"join(1..42,complement(346..723))",
	"join(complement(1..42),346..723)",
	"join(complement(1..42),complement(346..723))",

	// order
	"order(1..42,346..723)",
	"order(<1..42,346..723)",
	"order(1..>42,346..723)",
	"order(<1..>42,346..723)",
	"order(1..42,<346..723)",
	"order(1..42,346..>723)",
	"order(1..42,<346..>723)",
	"order(<1..42,<346..723)",
	"order(1..>42,346..>723)",
	"order(<1..>42,<346..>723)",
	"order(1..42,complement(346..723))",
	"order(complement(1..42),346..723)",
	"order(complement(1..42),complement(346..723))",
}

func TestLocationIO(t *testing.T) {
	for _, in := range locationTests {
		loc, err := AsLocation(in)
		if err != nil {
			t.Errorf("AsLocation(%q): %v", in, err)
		}
		out := loc.String()
		if out != in {
			t.Errorf("loc.String() = %q, want %q", out, in)
		}
	}

	loc0 := NewRangeLocation(0, 1)
	loc1 := NewPartialRangeLocation(0, 1, false, false)
	if !same(loc0, loc1) {
		t.Errorf("%#v != %#v", loc0, loc1)
	}

	if _, err := AsLocation(""); err == nil {
		t.Errorf("AsLocation(\"\") expected error")
	}
}
