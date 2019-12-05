package gt1_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
)

type Stringer string

func (s Stringer) String() string { return string(s) }

func TestSeq(t *testing.T) {
	s := "atgc"
	p := []byte(s)
	r := []rune(s)

	seqs := gt1.Seq(s)
	seqp := gt1.Seq(p)
	seqr := gt1.Seq(r)
	seqi := gt1.Seq(Stringer(s))
	seq := gt1.Seq(seqs)

	assert.Apply(t,
		assert.Equal(seqs, seq),
		assert.Equal(seqp, seq),
		assert.Equal(seqr, seq),
		assert.Equal(seqi, seq),

		assert.True(gt1.Equal(seqs, seq)),
		assert.True(gt1.Equal(seqp, seq)),
		assert.True(gt1.Equal(seqr, seq)),
		assert.True(gt1.Equal(seqi, seq)),

		assert.Panic(func() { gt1.Seq(0) }),
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
		cases[i] = assert.Equal(gt1.Skew(seq, nSet, pSet), skew)
	}

	assert.Apply(t, cases...)
}
