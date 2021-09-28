package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

type testSequence []byte

func (seq testSequence) Bytes() []byte {
	return Seq(seq).Bytes()
}

func (seq testSequence) Len() int {
	return len(seq)
}

func (seq testSequence) Slice(i, j int) Sequence {
	return Slice(Seq(seq), i, j)
}

func (seq testSequence) Insert(pos int, arg Sequence) Sequence {
	return Insert(Seq(seq), pos, arg)
}

func (seq testSequence) Replace(pos int, arg Sequence) Sequence {
	return Replace(Seq(seq), pos, arg)
}

func (seq testSequence) Delete(pos, arg int) Sequence {
	return Delete(Seq(seq), pos, arg)
}

var lenTests = []struct {
	in  Sequence
	out int
}{
	{AsSequence("atgc"), 4},
	{testSequence([]byte("atgc")), 4},
}

func TestLen(t *testing.T) {
	for i, tt := range lenTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := Len(tt.in)
			if out != tt.out {
				t.Errorf("Len(%q) = %d, want %d", string(tt.in.Bytes()), out, tt.out)
			}
		})
	}
}

var equalsTests = []struct {
	lhs, rhs Sequence
	out      bool
}{
	{AsSequence("atgc"), AsSequence("atgc"), true},
	{AsSequence("atgc"), AsSequence("cgta"), false},
}

func TestEquals(t *testing.T) {
	for i, tt := range equalsTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := Equals(tt.lhs, tt.rhs)
			if out != tt.out {
				t.Errorf(
					"Equals(%q, %q) = %t, want %t",
					string(tt.lhs.Bytes()), string(tt.rhs.Bytes()),
					out, tt.out,
				)
			}
		})
	}
}

var asSequenceTests = []struct {
	in  interface{}
	out Sequence
}{
	{testSequence("atgc"), Seq("atgc")}, // case 1
	{[]byte("atgc"), Seq("atgc")},       // case 2
	{[]rune("atgc"), Seq("atgc")},       // case 3
	{"atgc", Seq("atgc")},               // case 4
	{byte('a'), Seq("a")},               // case 5
	{rune('a'), Seq("a")},               // case 6
}

func TestAsSequence(t *testing.T) {
	for i, tt := range asSequenceTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := AsSequence(tt.in)
			if !Equals(out, tt.out) {
				t.Errorf(
					"New(%v) = %q, want %q",
					tt.in,
					string(tt.out.Bytes()),
					string(out.Bytes()),
				)
			}
		})
	}
	testutils.Panics(t, func() {
		AsSequence(0)
	})
}

var operationTests = []struct {
	op  Operation
	in  Sequence
	out Sequence
}{
	// case 1
	{
		Slicer(2, 6),
		Seq("atgcatgc"),
		Seq("gcat"),
	},

	// case 2
	{
		Inserter(2, Seq("gcat")),
		Seq("atgc"),
		Seq("atgcatgc"),
	},

	// case 3
	{
		Replacer(2, Seq("gcat")),
		Seq("atatgcgc"),
		Seq("atgcatgc"),
	},

	// case 4
	{
		Deleter(2, 4),
		Seq("atgcatgc"),
		Seq("atgc"),
	},

	// case 5
	{
		Slicer(2, 6),
		testSequence("atgcatgc"),
		Seq("gcat"),
	},

	// case 6
	{
		Inserter(2, Seq("gcat")),
		testSequence("atgc"),
		Seq("atgcatgc"),
	},

	// case 7
	{
		Replacer(2, Seq("gcat")),
		testSequence("atatgcgc"),
		Seq("atgcatgc"),
	},

	// case 8
	{
		Deleter(2, 4),
		testSequence("atgcatgc"),
		Seq("atgc"),
	},
}

func TestOperation(t *testing.T) {
	for i, tt := range operationTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := tt.op(tt.in)
			if !Equals(tt.out, out) {
				t.Errorf(
					"Op(%v) = %q, want %q",
					tt.in,
					string(tt.out.Bytes()),
					string(out.Bytes()),
				)
			}
		})
	}
}
