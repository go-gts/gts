package gts_test

import (
	"testing"

	"gopkg.in/ktnyt/assert.v1"
	"gopkg.in/ktnyt/gts.v0"
)

type Stringer string

func (s Stringer) String() string { return string(s) }

func TestSeq(t *testing.T) {
	s := "atgc"
	p := []byte(s)
	r := []rune(s)

	seqs := gts.Seq(s)
	seqp := gts.Seq(p)
	seqr := gts.Seq(r)
	seqi := gts.Seq(Stringer(s))
	seq := gts.Seq(seqs)

	assert.Apply(t,
		assert.Equal(seqs, seq),
		assert.Equal(seqp, seq),
		assert.Equal(seqr, seq),
		assert.Equal(seqi, seq),

		assert.True(gts.Equal(seqs, seq)),
		assert.True(gts.Equal(seqp, seq)),
		assert.True(gts.Equal(seqr, seq)),
		assert.True(gts.Equal(seqi, seq)),

		assert.Panic(func() { gts.Seq(0) }),
	)
}

func TestSlice(t *testing.T) {
	seq := gts.Seq("atatgcgc")
	e := gts.Seq("atgc")
	s := gts.Slice(seq, 2, 6)

	assert.Apply(t, assert.Equal(s, e))
}

func TestFragment(t *testing.T) {
	seq := gts.Seq("atgcatgc")

	e44 := gts.Seq("atgc")
	f44 := gts.Fragment(seq, 4, 4)

	e24 := gts.Seq("at")
	f24 := gts.Fragment(seq, 2, 4)

	e42 := []gts.Sequence{gts.Seq("atgc"), gts.Seq("gcat")}
	f42 := gts.Fragment(seq, 4, 2)

	assert.Apply(t,
		assert.Equal(f44[0], e44),
		assert.Equal(f44[1], e44),

		assert.Equal(f24[0], e24),
		assert.Equal(f24[1], e24),

		assert.Equal(f42[0], e42[0]),
		assert.Equal(f42[1], e42[1]),
	)
}

func TestComposition(t *testing.T) {
	seq := gts.Seq("atgcatgc")
	c := gts.Composition(seq)
	e := map[byte]int{'a': 2, 't': 2, 'g': 2, 'c': 2}

	assert.Apply(t, assert.Equal(c, e))
}

func TestSkew(t *testing.T) {
	seq := gts.Seq("atgcatgc")

	values := []struct {
		NSet string
		PSet string
		Skew float64
	}{
		{"g", "c", 0.0},
		{"a", "t", 0.0},
		{"g", "", -1.0},
		{"", "g", 1.0},
	}

	cases := make([]assert.F, len(values))
	for i, value := range values {
		nSet, pSet, skew := value.NSet, value.PSet, value.Skew
		cases[i] = assert.Equal(gts.Skew(seq, nSet, pSet), skew)
	}

	assert.Apply(t, cases...)
}
