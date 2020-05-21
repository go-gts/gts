package gts

import (
	"testing"

	"github.com/go-gts/gts/testutils"
)

type LenObj []byte

func (obj LenObj) Info() interface{} {
	return nil
}

func (obj LenObj) Features() FeatureTable {
	return nil
}

func (obj LenObj) Bytes() []byte {
	return obj
}

func (obj LenObj) Len() int {
	return len(obj)
}

func TestSequence(t *testing.T) {
	info := "test sequence"
	p := []byte("atgc")
	seq := New(info, nil, p)

	testutils.Equals(t, seq.Info(), info)
	testutils.Equals(t, seq.Bytes(), p)

	if Len(seq) != Len(LenObj(p)) {
		t.Errorf("Len(seq) = %d, want %d", Len(seq), len(p))
	}
}

func TestSlice(t *testing.T) {
	p := []byte("atatexpcgc")
	in := New(nil, nil, p)
	for i := 0; i < len(p); i++ {
		for j := i; j < len(p); j++ {
			out, exp := Slice(in, i, j), New(nil, nil, p[i:j])
			if !Equal(out, exp) {
				t.Errorf(
					"Slice(%q, %d, %d) = %q, want %q",
					string(in.Bytes()), i, j,
					string(out.Bytes()),
					string(exp.Bytes()),
				)
			}
		}
	}
}

func TestReverse(t *testing.T) {
	in, exp := New(nil, nil, []byte("atgc")), New(nil, nil, []byte("cgta"))
	out := Reverse(in)
	if !Equal(out, exp) {
		t.Errorf(
			"Reverse(%q) = %q, want %q",
			string(in.Bytes()),
			string(out.Bytes()),
			string(exp.Bytes()),
		)
	}
}
