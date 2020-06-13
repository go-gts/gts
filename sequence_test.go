package gts

import (
	"strings"
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

	cpy := Copy(seq)

	testutils.Equals(t, seq.Info(), cpy.Info())
	testutils.Equals(t, seq.Bytes(), cpy.Bytes())

	if Len(seq) != Len(LenObj(p)) {
		t.Errorf("Len(seq) = %d, want %d", Len(seq), len(p))
	}
}

func TestSlice(t *testing.T) {
	p := []byte(strings.Repeat("atgc", 2))
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	loc := Range(0, len(p))
	ff := []Feature{
		{
			Key:        "source",
			Location:   loc,
			Qualifiers: qfs,
		},
	}
	info := "info"
	in := New(info, ff, p)

	for i := 0; i < len(p); i++ {
		for j := i; j < len(p); j++ {
			out, exp := Slice(in, i, j), New(info, ff, p[i:j])
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

func TestConcat(t *testing.T) {
	out := Concat()
	exp := New(nil, nil, nil)
	if !Equal(out, exp) {
		t.Errorf("Concat() = %v, want %v", out, exp)
	}

	p := []byte(strings.Repeat("atgc", 2))
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	loc := Range(0, len(p))
	f := Feature{
		Key:        "source",
		Location:   loc,
		Qualifiers: qfs,
	}

	ff := []Feature{f}
	info := "info"
	seq := New(info, ff, p)

	out = Concat(seq)
	exp = seq
	if !Equal(out, exp) {
		t.Errorf("Concat() = %v, want %v", out, exp)
	}

	out = Concat(seq, seq)
	g := Feature{
		Key:        f.Key,
		Location:   f.Location.Shift(Span{0, Len(seq)}),
		Qualifiers: qfs,
	}
	exp = New(info, append(ff, g), append(p, p...))
	if !Equal(out, exp) {
		t.Errorf("Concat() = %v, want %v", out, exp)
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

func TestWith(t *testing.T) {
	p := []byte(strings.Repeat("atgc", 100))
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	loc := Range(0, len(p))
	ff := []Feature{
		{
			Key:        "source",
			Location:   loc,
			Qualifiers: qfs,
		},
	}
	info := "info"

	in := New(nil, nil, nil)
	out := WithInfo(in, info)
	testutils.Equals(t, out, New(info, nil, nil))

	out = WithFeatures(in, ff)
	testutils.Equals(t, out, New(nil, ff, nil))

	out = WithBytes(in, p)
	testutils.Equals(t, out, New(nil, nil, p))
}

type withTest struct {
	info  interface{}
	table FeatureTable
	data  []byte
}

func newWithTest(info interface{}, table FeatureTable, p []byte) withTest {
	return withTest{info, table, p}
}

func (wt withTest) Info() interface{} {
	return wt.info
}

func (wt withTest) Features() FeatureTable {
	return wt.table
}

func (wt withTest) Bytes() []byte {
	return wt.data
}

func (wt withTest) WithInfo(info interface{}) Sequence {
	return withTest{info, wt.table, wt.data}
}

func (wt withTest) WithFeatures(ff FeatureTable) Sequence {
	return withTest{wt.info, ff, wt.data}
}

func (wt withTest) WithBytes(p []byte) Sequence {
	return withTest{wt.info, wt.table, p}
}

func TestWithInterface(t *testing.T) {
	p := []byte(strings.Repeat("atgc", 100))
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	loc := Range(0, len(p))
	ff := []Feature{
		{
			Key:        "source",
			Location:   loc,
			Qualifiers: qfs,
		},
	}
	info := "info"

	in := newWithTest(nil, nil, nil)
	out := WithInfo(in, info)
	testutils.Equals(t, out, newWithTest(info, nil, nil))

	out = WithFeatures(in, ff)
	testutils.Equals(t, out, newWithTest(nil, ff, nil))

	out = WithBytes(in, p)
	testutils.Equals(t, out, newWithTest(nil, nil, p))
}
