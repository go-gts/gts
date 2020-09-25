package gts

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
)

var regionResizeTests = []struct {
	in       Region
	modifier Modifier
	out      Region
}{
	{Forward{3, 6}, Head(+0), Forward{3, 3}},
	{Forward{3, 6}, Head(+1), Forward{4, 4}},
	{Forward{3, 6}, Head(-1), Forward{2, 2}},

	{Forward{3, 6}, Tail(+0), Forward{6, 6}},
	{Forward{3, 6}, Tail(+1), Forward{7, 7}},
	{Forward{3, 6}, Tail(-1), Forward{5, 5}},

	{Forward{3, 6}, HeadTail{+0, +0}, Forward{3, 6}},
	{Forward{3, 6}, HeadTail{+0, +1}, Forward{3, 7}},
	{Forward{3, 6}, HeadTail{+2, +0}, Forward{5, 6}},
	{Forward{3, 6}, HeadTail{+0, -1}, Forward{3, 5}},
	{Forward{3, 6}, HeadTail{-2, +0}, Forward{1, 6}},
	{Forward{3, 6}, HeadTail{+2, -1}, Forward{5, 5}},
	{Forward{3, 6}, HeadTail{-2, +1}, Forward{1, 7}},
	{Forward{3, 6}, HeadTail{+2, +1}, Forward{5, 7}},
	{Forward{3, 6}, HeadTail{-2, -1}, Forward{1, 5}},

	{Forward{3, 6}, HeadHead{+0, +0}, Forward{3, 3}},
	{Forward{3, 6}, HeadHead{+0, +1}, Forward{3, 4}},
	{Forward{3, 6}, HeadHead{+2, +0}, Forward{5, 5}},
	{Forward{3, 6}, HeadHead{+0, -1}, Forward{3, 3}},
	{Forward{3, 6}, HeadHead{-2, +0}, Forward{1, 3}},
	{Forward{3, 6}, HeadHead{+2, -1}, Forward{5, 5}},
	{Forward{3, 6}, HeadHead{-2, +1}, Forward{1, 4}},
	{Forward{3, 6}, HeadHead{+2, +1}, Forward{5, 5}},
	{Forward{3, 6}, HeadHead{-2, -1}, Forward{1, 2}},

	{Forward{3, 6}, TailTail{+0, +0}, Forward{6, 6}},
	{Forward{3, 6}, TailTail{+0, +1}, Forward{6, 7}},
	{Forward{3, 6}, TailTail{+2, +0}, Forward{8, 8}},
	{Forward{3, 6}, TailTail{+0, -1}, Forward{6, 6}},
	{Forward{3, 6}, TailTail{-2, +0}, Forward{4, 6}},
	{Forward{3, 6}, TailTail{+2, -1}, Forward{8, 8}},
	{Forward{3, 6}, TailTail{-2, +1}, Forward{4, 7}},
	{Forward{3, 6}, TailTail{+2, +1}, Forward{8, 8}},
	{Forward{3, 6}, TailTail{-2, -1}, Forward{4, 5}},

	{Backward{3, 6}, Head(+0), Backward{6, 6}},
	{Backward{3, 6}, Head(+1), Backward{5, 5}},
	{Backward{3, 6}, Head(-1), Backward{7, 7}},

	{Backward{3, 6}, Tail(+0), Backward{3, 3}},
	{Backward{3, 6}, Tail(+1), Backward{2, 2}},
	{Backward{3, 6}, Tail(-1), Backward{4, 4}},

	{Backward{3, 6}, HeadTail{+0, +0}, Backward{3, 6}},
	{Backward{3, 6}, HeadTail{+0, +1}, Backward{2, 6}},
	{Backward{3, 6}, HeadTail{+2, +0}, Backward{3, 4}},
	{Backward{3, 6}, HeadTail{+0, -1}, Backward{4, 6}},
	{Backward{3, 6}, HeadTail{-2, +0}, Backward{3, 8}},
	{Backward{3, 6}, HeadTail{+2, -1}, Backward{4, 4}},
	{Backward{3, 6}, HeadTail{-2, +1}, Backward{2, 8}},
	{Backward{3, 6}, HeadTail{+2, +1}, Backward{2, 4}},
	{Backward{3, 6}, HeadTail{-2, -1}, Backward{4, 8}},

	{Backward{3, 6}, HeadHead{+0, +0}, Backward{6, 6}},
	{Backward{3, 6}, HeadHead{+0, +1}, Backward{5, 6}},
	{Backward{3, 6}, HeadHead{+2, +0}, Backward{4, 4}},
	{Backward{3, 6}, HeadHead{+0, -1}, Backward{6, 6}},
	{Backward{3, 6}, HeadHead{-2, +0}, Backward{6, 8}},
	{Backward{3, 6}, HeadHead{+2, -1}, Backward{4, 4}},
	{Backward{3, 6}, HeadHead{-2, +1}, Backward{5, 8}},
	{Backward{3, 6}, HeadHead{+2, +1}, Backward{4, 4}},
	{Backward{3, 6}, HeadHead{-2, -1}, Backward{7, 8}},

	{Backward{3, 6}, TailTail{+0, +0}, Backward{3, 3}},
	{Backward{3, 6}, TailTail{+0, +1}, Backward{2, 3}},
	{Backward{3, 6}, TailTail{+2, +0}, Backward{1, 1}},
	{Backward{3, 6}, TailTail{+0, -1}, Backward{3, 3}},
	{Backward{3, 6}, TailTail{-2, +0}, Backward{3, 5}},
	{Backward{3, 6}, TailTail{+2, -1}, Backward{1, 1}},
	{Backward{3, 6}, TailTail{-2, +1}, Backward{2, 5}},
	{Backward{3, 6}, TailTail{+2, +1}, Backward{1, 1}},
	{Backward{3, 6}, TailTail{-2, -1}, Backward{4, 5}},

	{Regions{Forward{3, 6}, Forward{13, 16}}, Head(+0), Forward{3, 3}},
	{Regions{Forward{3, 6}, Forward{13, 16}}, Head(+7), Forward{17, 17}},
	{Regions{Forward{3, 6}, Forward{13, 16}}, Tail(+0), Forward{16, 16}},
	{Regions{Forward{3, 6}, Forward{13, 16}}, Tail(-7), Forward{2, 2}},
	{Regions{Forward{13, 16}, Forward{3, 6}}, Head(+0), Forward{13, 13}},
	{Regions{Forward{13, 16}, Forward{3, 6}}, Head(+7), Forward{7, 7}},
	{Regions{Forward{13, 16}, Forward{3, 6}}, Tail(+0), Forward{6, 6}},
	{Regions{Forward{13, 16}, Forward{3, 6}}, Tail(-7), Forward{12, 12}},

	{Regions{Forward{3, 6}, Forward{13, 16}}, HeadTail{0, 0}, Regions{Forward{3, 6}, Forward{13, 16}}},
	{Regions{Forward{3, 6}, Forward{13, 16}}, HeadTail{4, -4}, Forward{14, 14}},
	{Regions{Forward{3, 6}, Forward{13, 16}}, HeadHead{-2, 4}, Regions{Forward{1, 6}, Forward{13, 14}}},
	{Regions{Forward{3, 6}, Forward{13, 16}}, TailTail{-4, 2}, Regions{Forward{5, 6}, Forward{13, 18}}},
}

func TestRegionResize(t *testing.T) {
	for i, tt := range regionResizeTests {
		out := tt.in.Resize(tt.modifier)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf(
				"\ncase [%d]:\n  resize by %s\n   in: %#v\n  out: %#v\n  exp: %#v",
				i+1, tt.modifier, tt.in, out, tt.out,
			)
		}
	}
}

var regionLocateTests = []struct {
	in  Region
	out Sequence
}{
	{Forward{2, 6}, New(nil, nil, []byte("gcat"))},
	{Backward{2, 6}, New(nil, nil, []byte("atgc"))},
	{Regions{Forward{0, 2}, Forward{4, 6}}, New(nil, nil, []byte("atat"))},
}

func TestRegionLocate(t *testing.T) {
	seq := New(nil, nil, []byte("atgcatgc"))
	for _, tt := range regionLocateTests {
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
		if cmp.Len() != tt.in.Len() {
			t.Errorf("%s.Len() = %d, want %d", cmp, cmp.Len(), tt.in.Len())
		}
		if !reflect.DeepEqual(cmp.Complement(), tt.in) {
			t.Errorf(
				"%s.Complement() = %s, want %s",
				cmp, cmp.Complement(), tt.in,
			)
		}
		out = cmp.Locate(seq)
		exp = Reverse(Complement(tt.out))
		if !Equal(out, exp) {
			t.Errorf(
				"%s.Locate(%q) = %q, want %q",
				cmp, string(seq.Bytes()),
				string(out.Bytes()), string(exp.Bytes()),
			)
		}
	}
}

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
