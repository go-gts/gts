package gts

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-gts/gts/testutils"
	"github.com/go-test/deep"
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

func TestInsert(t *testing.T) {
	p := []byte("atgcatgc")
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	ff := []Feature{{"source", Range(0, len(p)), qfs, nil}}
	info := "info"
	in := New(info, ff, p)
	out := Insert(in, 2, in)

	q := []byte("atatgcatgcgcatgc")
	gg := []Feature{
		{"source", Join(Range(0, 2), Range(2+len(p), len(q))), qfs, nil},
		{"source", Range(2, 2+len(p)), qfs, nil},
	}
	exp := New(info, gg, q)

	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Insert(seq, 2, seq).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if diff := deep.Equal(out.Features(), exp.Features()); diff != nil {
		t.Errorf("Insert(seq, 2, seq).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if diff := deep.Equal(out.Bytes(), exp.Bytes()); diff != nil {
		t.Errorf("Insert(seq, 2, seq).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestEmbed(t *testing.T) {
	p := []byte("atgcatgc")
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	ff := []Feature{{"source", Range(0, len(p)), qfs, nil}}
	info := "info"
	in := New(info, ff, p)
	out := Embed(in, 2, in)

	q := []byte("atatgcatgcgcatgc")
	gg := []Feature{
		{"source", Range(0, len(q)), qfs, nil},
		{"source", Range(2, 2+len(p)), qfs, nil},
	}
	exp := New(info, gg, q)

	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Embed(seq, 2, seq).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if diff := deep.Equal(out.Features(), exp.Features()); diff != nil {
		t.Errorf("Embed(seq, 2, seq).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if diff := deep.Equal(out.Bytes(), exp.Bytes()); diff != nil {
		t.Errorf("Embed(seq, 2, seq).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestDelete(t *testing.T) {
	p := []byte("atgcatgc")
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	ff := []Feature{{"source", Range(0, len(p)), qfs, nil}}
	info := "info"
	in := New(info, ff, p)
	out := Delete(in, 3, 2)

	q := []byte("atgtgc")
	gg := []Feature{{"source", Range(0, len(q)), qfs, nil}}
	exp := New(info, gg, q)

	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Delete(seq, 2, seq).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if diff := deep.Equal(out.Features(), exp.Features()); diff != nil {
		t.Errorf("Delete(seq, 2, seq).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if diff := deep.Equal(out.Bytes(), exp.Bytes()); diff != nil {
		t.Errorf("Delete(seq, 2, seq).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestSlice(t *testing.T) {
	p := []byte("atgcatgc")
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	ff := []Feature{{"source", Range(0, len(p)), qfs, nil}, {"gene", Range(3, 5), qfs, nil}}
	info := "info"
	in := New(info, ff, p)

	gg := []Feature{{"source", Range(0, 4), qfs, nil}, {"gene", Range(1, 3), qfs, nil}}
	out, exp := Slice(in, 2, 6), New(info, gg, p[2:6])
	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Slice(in, %d, %d).Info() = %v, want %v", 2, 6, out.Info(), exp.Info())
	}
	if diff := deep.Equal(out.Features(), exp.Features()); diff != nil {
		t.Errorf("Slice(in, %d, %d).Features() = %v, want %v", 2, 6, out.Features(), exp.Features())
	}
	if diff := deep.Equal(out.Bytes(), exp.Bytes()); diff != nil {
		t.Errorf("Slice(in, %d, %d).Bytes() = %v, want %v", 2, 6, out.Bytes(), exp.Bytes())
	}
}

func TestConcat(t *testing.T) {
	out := Concat()
	exp := New(nil, nil, nil)
	if !Equal(out, exp) {
		t.Errorf("Concat() = %v, want %v", out, exp)
	}

	p := []byte("atgcatgc")
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
		Location:   f.Location.Shift(0, Len(seq), false),
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
