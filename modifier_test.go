package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

var asModifierTests = []struct {
	in  string
	out Modifier
}{
	{"^", Head(0)},      // case 1
	{"^+42", Head(42)},  // case 2
	{"^-42", Head(-42)}, // case 3

	{"$", Tail(0)},      // case 4
	{"$+42", Tail(42)},  // case 5
	{"$-42", Tail(-42)}, // case 6

	{"^..$", HeadTail{0, 0}},       // case 7
	{"^+1..$+1", HeadTail{+1, +1}}, // case 8
	{"^-1..$+1", HeadTail{-1, +1}}, // case 9
	{"^-1..$-1", HeadTail{-1, -1}}, // case 10

	{"^..^", HeadHead{0, 0}},       // case 11
	{"^+1..^+1", HeadHead{+1, +1}}, // case 12
	{"^-1..^+1", HeadHead{-1, +1}}, // case 13
	{"^-1..^-1", HeadHead{-1, -1}}, // case 14

	{"$..$", TailTail{0, 0}},       // case 15
	{"$+1..$+1", TailTail{+1, +1}}, // case 16
	{"$-1..$+1", TailTail{-1, +1}}, // case 17
	{"$-1..$-1", TailTail{-1, -1}}, // case 18
}

var asModifierFailTests = []string{
	"",       // case 1
	"^-2..0", // case 2
	"$..^",   // case 3
}

func TestAsModifier(t *testing.T) {
	for i, tt := range asModifierTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out, err := AsModifier(tt.in)
			if err != nil {
				t.Errorf("AsModifier(%q): %v", tt.in, err)
				return
			}
			if out.String() != tt.out.String() {
				t.Errorf("AsModifier(%q) = %q, want %q", tt.in, out, tt.out)
			}
		})
	}

	for i, in := range asModifierFailTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			if _, err := AsModifier(in); err == nil {
				t.Errorf("expected error in AsModifier(%q)", in)
			}
		})
	}
}
