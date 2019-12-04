package gt1_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
)

func TestSeq(t *testing.T) {
	s := "atgc"
	p := []byte(s)
	r := []rune(s)

	seqs := gt1.Seq(s)
	seqp := gt1.Seq(p)
	seqr := gt1.Seq(r)
	seq := gt1.Seq(seqs)

	assert.Apply(t,
		assert.Equal(seq, seqs),
		assert.Equal(seq, seqp),
		assert.Equal(seq, seqr),
		assert.True(gt1.Equal(seq, seqs)),
		assert.True(gt1.Equal(seq, seqp)),
		assert.True(gt1.Equal(seq, seqr)),
	)
}

func TestSlice(t *testing.T) {
	seq := gt1.Seq("atatgcgc")
	e := gt1.Seq("atgc")
	s := gt1.Slice(seq, 2, 6)

	assert.Apply(t, assert.Equal(s, e))
}

func TestFragment(t *testing.T) {
	seq := gt1.Seq("atgcatgc")

	e44 := gt1.Seq("atgc")
	f44 := gt1.Fragment(seq, 4, 4)

	e24 := gt1.Seq("at")
	f24 := gt1.Fragment(seq, 2, 4)

	e42 := []gt1.Sequence{gt1.Seq("atgc"), gt1.Seq("gcat")}
	f42 := gt1.Fragment(seq, 4, 2)

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
	seq := gt1.Seq("atgcatgc")
	c := gt1.Composition(seq)
	e := map[byte]int{'a': 2, 't': 2, 'g': 2, 'c': 2}

	assert.Apply(t, assert.Equal(c, e))
}

func TestSkew(t *testing.T) {
	seq := gt1.Seq("atgcatgc")

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

	cases := make([]assert.F, len(values))
	for i, value := range values {
		nSet, pSet, skew := value.nSet, value.pSet, value.skew
		cases[i] = assert.Equal(gt1.Skew(seq, nSet, pSet), skew)
	}

	assert.Apply(t, cases...)
}
