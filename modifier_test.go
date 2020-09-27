package gts

import (
	"testing"
)

var asModifierTests = []struct {
	in  string
	out Modifier
}{
	{"^", Head(0)},
	{"^+42", Head(42)},
	{"^-42", Head(-42)},

	{"$", Tail(0)},
	{"$+42", Tail(42)},
	{"$-42", Tail(-42)},

	{"^..$", HeadTail{0, 0}},
	{"^+1..$+1", HeadTail{+1, +1}},
	{"^-1..$+1", HeadTail{-1, +1}},
	{"^-1..$-1", HeadTail{-1, -1}},

	{"^..^", HeadHead{0, 0}},
	{"^+1..^+1", HeadHead{+1, +1}},
	{"^-1..^+1", HeadHead{-1, +1}},
	{"^-1..^-1", HeadHead{-1, -1}},

	{"$..$", TailTail{0, 0}},
	{"$+1..$+1", TailTail{+1, +1}},
	{"$-1..$+1", TailTail{-1, +1}},
	{"$-1..$-1", TailTail{-1, -1}},
}

var asModifierFailTests = []string{
	"",
	"^-2..0",
	"$..^",
}

func TestAsModifier(t *testing.T) {
	for _, tt := range asModifierTests {
		out, err := AsModifier(tt.in)
		if err != nil {
			t.Errorf("AsModifier(%q): %v", tt.in, err)
			continue
		}
		if out.String() != tt.out.String() {
			t.Errorf("AsModifier(%q) = %q, want %q", tt.in, out, tt.out)
		}
	}

	for _, in := range asModifierFailTests {
		if _, err := AsModifier(in); err == nil {
			t.Errorf("expected error in AsModifier(%q)", in)
		}
	}
}
