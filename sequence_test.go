package gts

import (
	"testing"
)

type Stringer string

func (s Stringer) String() string { return string(s) }

func TestSeq(t *testing.T) {
	s := "atgc"
	p := []byte(s)
	r := []rune(s)

	seqs := Seq(s)
	seqp := Seq(p)
	seqr := Seq(r)
	seqi := Seq(Stringer(s))
	seq := Seq(seqs)

	equals(t, seqs, seq)
	equals(t, seqp, seq)
	equals(t, seqr, seq)
	equals(t, seqi, seq)

	equals(t, Equal(seqs, seq), true)
	equals(t, Equal(seqp, seq), true)
	equals(t, Equal(seqr, seq), true)
	equals(t, Equal(seqi, seq), true)

	PanicTest(t, func(t *testing.T) {
		t.Helper()
		Seq(0)
	})
}

func TestSlice(t *testing.T) {
	seq := Seq("atatgcgc")
	e := Seq("atgc")
	s := Slice(seq, 2, 6)

	equals(t, s, e)
}

func TestFragment(t *testing.T) {
	seq := Seq("atgcatgc")

	e44 := Seq("atgc")
	f44 := Fragment(seq, 4, 4)

	e24 := Seq("at")
	f24 := Fragment(seq, 2, 4)

	e42 := []Sequence{Seq("atgc"), Seq("gcat")}
	f42 := Fragment(seq, 4, 2)

	equals(t, f44[0], e44)
	equals(t, f44[1], e44)

	equals(t, f24[0], e24)
	equals(t, f24[1], e24)

	equals(t, f42[0], e42[0])
	equals(t, f42[1], e42[1])
}

func TestComposition(t *testing.T) {
	seq := Seq("atgcatgc")
	c := Composition(seq)
	e := map[byte]int{'a': 2, 't': 2, 'g': 2, 'c': 2}

	equals(t, c, e)
}

func TestSkew(t *testing.T) {
	seq := Seq("atgcatgc")

	values := []struct {
		nSet string
		pSet string
		skew float64
	}{
		{"g", "c", 0.0},
		{"a", "t", 0.0},
		{"g", "", -1.0},
		{"", "g", 1.0},
	}

	for _, value := range values {
		nSet, pSet, skew := value.nSet, value.pSet, value.skew
		out := Skew(seq, nSet, pSet)
		if out != skew {
			t.Errorf("Skew(%q, %q, %q) = %f, want %f", seq, nSet, pSet, out, skew)
		}
	}
}
