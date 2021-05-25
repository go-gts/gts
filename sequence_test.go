package gts

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func featuresEqual(a, b []Feature) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

func bytesEqual(a, b []byte) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

type infoInterface [2]int

func (info infoInterface) Shift(i, n int) interface{} {
	return infoInterface{i, n}
}

func (info infoInterface) Expand(i, n int) interface{} {
	return infoInterface{i, n}
}

func (info infoInterface) Slice(start, end int) interface{} {
	return infoInterface{start, end}
}

func TestShiftableExpandable(t *testing.T) {
	var in, out interface{}
	i, n := 3, 6
	exp := infoInterface{i, n}

	in = infoInterface{0, 0}
	out = tryShift(in, i, n)
	if !reflect.DeepEqual(out, exp) {
		t.Errorf("tryShift(%v, %d, %d) = %v, want %v", in, i, n, out, exp)
	}

	in = infoInterface{0, 0}
	out = tryExpand(in, i, n)
	if !reflect.DeepEqual(out, exp) {
		t.Errorf("tryExpand(%v, %d, %d) = %v, want %v", in, i, n, out, exp)
	}

	in = infoInterface{0, 0}
	out = trySlice(in, i, n)
	if !reflect.DeepEqual(out, exp) {
		t.Errorf("trySlice(%v, %d, %d) = %v, want %v", in, i, n, out, exp)
	}

	in = "info"
	out = tryShift(in, i, n)
	if !reflect.DeepEqual(out, in) {
		t.Errorf("tryShift(%v, %d, %d) = %v, want %v", in, i, n, out, exp)
	}

	in = "info"
	out = tryExpand(in, i, n)
	if !reflect.DeepEqual(out, in) {
		t.Errorf("tryExpand(%v, %d, %d) = %v, want %v", in, i, n, out, exp)
	}

	in = "info"
	out = trySlice(in, i, n)
	if !reflect.DeepEqual(out, in) {
		t.Errorf("trySlice(%v, %d, %d) = %v, want %v", in, i, n, out, exp)
	}
}

type LenObj []byte

func (obj LenObj) Info() interface{} {
	return nil
}

func (obj LenObj) Features() FeatureSlice {
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
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{NewFeature("source", Range(0, len(p)), props)}
	info := "info"
	in := New(info, ff, p)
	out := Insert(in, 2, in)

	q := []byte("atatgcatgcgcatgc")
	gg := []Feature{
		NewFeature("source", Join(Range(0, 2), Range(2+len(p), len(q))), props),
		NewFeature("source", Range(2, 2+len(p)), props),
	}
	exp := New(info, gg, q)

	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Insert(seq, 2, seq).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if !featuresEqual(out.Features(), exp.Features()) {
		t.Errorf("Insert(seq, 2, seq).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if !bytesEqual(out.Bytes(), exp.Bytes()) {
		t.Errorf("Insert(seq, 2, seq).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestEmbed(t *testing.T) {
	p := []byte("atgcatgc")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{NewFeature("source", Range(0, len(p)), props)}
	info := "info"
	in := New(info, ff, p)
	out := Embed(in, 2, in)

	q := []byte("atatgcatgcgcatgc")
	gg := []Feature{
		NewFeature("source", Range(0, len(q)), props),
		NewFeature("source", Range(2, 2+len(p)), props),
	}
	exp := New(info, gg, q)

	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Embed(seq, 2, seq).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if !featuresEqual(out.Features(), exp.Features()) {
		t.Errorf("Embed(seq, 2, seq).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if !bytesEqual(out.Bytes(), exp.Bytes()) {
		t.Errorf("Embed(seq, 2, seq).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestDelete(t *testing.T) {
	p := []byte("atgcatgc")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(4, 5), props),
	}
	info := "info"
	in := New(info, ff, p)
	out := Delete(in, 3, 4)

	q := []byte("atgc")
	gg := []Feature{
		NewFeature("source", Range(0, len(q)), props),
		NewFeature("gene", Between(3), props),
	}
	exp := New(info, gg, q)

	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Delete(seq, 3, 4).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if !featuresEqual(out.Features(), exp.Features()) {
		t.Errorf("Delete(seq, 3, 4).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if !bytesEqual(out.Bytes(), exp.Bytes()) {
		t.Errorf("Delete(seq, 3, 4).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestErase(t *testing.T) {
	p := []byte("atgcatgc")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(4, 5), props),
	}
	info := "info"
	in := New(info, ff, p)
	out := Erase(in, 3, 4)

	q := []byte("atgc")
	gg := []Feature{
		NewFeature("source", Range(0, len(q)), props),
	}
	exp := New(info, gg, q)

	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Erase(seq, 3, 4).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if !featuresEqual(out.Features(), exp.Features()) {
		t.Errorf("Erase(seq, 3, 4).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if !bytesEqual(out.Bytes(), exp.Bytes()) {
		t.Errorf("Erase(seq, 3, 4).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestSlice(t *testing.T) {
	p := []byte("atgcatgc")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(3, 5), props),
	}
	info := "info"
	in := New(info, ff, p)

	t.Run("Forward", func(t *testing.T) {
		gg := []Feature{
			NewFeature("source", Range(0, 4), props),
			NewFeature("gene", Range(1, 3), props),
		}
		out, exp := Slice(in, 2, 6), New(info, gg, p[2:6])
		if !reflect.DeepEqual(out.Info(), exp.Info()) {
			t.Errorf("Slice(in, %d, %d).Info() = %v, want %v", 2, 6, out.Info(), exp.Info())
		}
		if !featuresEqual(out.Features(), exp.Features()) {
			t.Errorf("Slice(in, %d, %d).Features() = %v, want %v", 2, 6, out.Features(), exp.Features())
		}
		if !bytesEqual(out.Bytes(), exp.Bytes()) {
			t.Errorf("Slice(in, %d, %d).Bytes() = %v, want %v", 2, 6, out.Bytes(), exp.Bytes())
		}
	})

	t.Run("Backward", func(t *testing.T) {
		gg := []Feature{NewFeature("source", Range(0, 4), props)}
		out, exp := Slice(in, 6, 2), New(info, gg, append(p[6:], p[:2]...))
		if !reflect.DeepEqual(out.Info(), exp.Info()) {
			t.Errorf("Slice(in, %d, %d).Info() = %v, want %v", 6, 2, out.Info(), exp.Info())
		}
		if !featuresEqual(out.Features(), exp.Features()) {
			t.Errorf("Slice(in, %d, %d).Features() = %v, want %v", 6, 2, out.Features(), exp.Features())
		}
		if !bytesEqual(out.Bytes(), exp.Bytes()) {
			t.Errorf("Slice(in, %d, %d).Bytes() = %v, want %v", 6, 2, out.Bytes(), exp.Bytes())
		}
	})

	t.Run("Negative", func(t *testing.T) {
		gg := []Feature{
			NewFeature("source", Range(0, 4), props),
			NewFeature("gene", Range(1, 3), props),
		}
		out, exp := Slice(in, -6, -2), New(info, gg, p[2:6])
		if !reflect.DeepEqual(out.Info(), exp.Info()) {
			t.Errorf("Slice(in, %d, %d).Info() = %v, want %v", -6, -2, out.Info(), exp.Info())
		}
		if !featuresEqual(out.Features(), exp.Features()) {
			t.Errorf("Slice(in, %d, %d).Features() = %v, want %v", -6, -2, out.Features(), exp.Features())
		}
		if !bytesEqual(out.Bytes(), exp.Bytes()) {
			t.Errorf("Slice(in, %d, %d).Bytes() = %v, want %v", -6, -2, out.Bytes(), exp.Bytes())
		}
	})
}

func TestConcat(t *testing.T) {
	out := Concat()
	exp := New(nil, nil, nil)
	if !Equal(out, exp) {
		t.Errorf("Concat() = %v, want %v", out, exp)
	}

	p := []byte("atgcatgc")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	f := NewFeature("source", Range(0, len(p)), props)

	ff := []Feature{f}
	info := "info"
	seq := New(info, ff, p)

	out = Concat(seq)
	exp = seq
	if !Equal(out, exp) {
		t.Errorf("Concat() = %v, want %v", out, exp)
	}

	out = Concat(seq, seq)
	g := NewFeature(f.Key, f.Loc.Shift(0, Len(seq)), f.Props.Clone())
	exp = New(info, append(ff, g), append(p, p...))
	if !Equal(out, exp) {
		t.Errorf("Concat() = %v, want %v", out, exp)
	}
}

func TestReverse(t *testing.T) {
	p, q := []byte("atgcatgc"), []byte("cgtacgta")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(2, 4), props),
	}
	gg := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(4, 6), props),
	}

	info := "info"

	in, exp := New(info, ff, p), New(info, gg, q)
	out := Reverse(in)
	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Reverse(in).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if !featuresEqual(out.Features(), exp.Features()) {
		t.Errorf("Reverse(in).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if !bytesEqual(out.Bytes(), exp.Bytes()) {
		t.Errorf("Reverse(in).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestRotate(t *testing.T) {
	p, q := []byte("aattggcc"), []byte("ccaattgg")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(2, 4), props),
	}
	gg := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(4, 6), props),
	}

	info := "info"

	in, exp := New(info, ff, p), New(info, gg, q)
	out := Rotate(in, -6)
	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Rotate(in, -6).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if !featuresEqual(out.Features(), exp.Features()) {
		t.Errorf("Rotate(in, -6).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if !bytesEqual(out.Bytes(), exp.Bytes()) {
		t.Errorf("Rotate(in, -6).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}

	out = Rotate(in, 10)
	if !reflect.DeepEqual(out.Info(), exp.Info()) {
		t.Errorf("Rotate(in, 10).Info() = %v, want %v", out.Info(), exp.Info())
	}
	if !featuresEqual(out.Features(), exp.Features()) {
		t.Errorf("Rotate(in, 10).Features() = %v, want %v", out.Features(), exp.Features())
	}
	if !bytesEqual(out.Bytes(), exp.Bytes()) {
		t.Errorf("Rotate(in, 10).Bytes() = %v, want %v", out.Bytes(), exp.Bytes())
	}
}

func TestSeqWith(t *testing.T) {
	p := []byte(strings.Repeat("atgc", 100))
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{NewFeature("source", Range(0, len(p)), props)}

	info := "info"

	in := New(nil, nil, nil)
	out := WithInfo(in, info)
	testutils.Equals(t, out, New(info, nil, nil))

	out = WithFeatures(in, ff)
	testutils.Equals(t, out, New(nil, ff, nil))

	out = WithBytes(in, p)
	testutils.Equals(t, out, New(nil, nil, p))
}

type seqWithTest struct {
	info  interface{}
	feats FeatureSlice
	data  []byte
}

func newSeqWithTest(info interface{}, ff FeatureSlice, p []byte) seqWithTest {
	return seqWithTest{info, ff, p}
}

func (wt seqWithTest) Info() interface{} {
	return wt.info
}

func (wt seqWithTest) Features() FeatureSlice {
	return wt.feats
}

func (wt seqWithTest) Bytes() []byte {
	return wt.data
}

func (wt seqWithTest) WithInfo(info interface{}) Sequence {
	return seqWithTest{info, wt.feats, wt.data}
}

func (wt seqWithTest) WithFeatures(ff []Feature) Sequence {
	return seqWithTest{wt.info, ff, wt.data}
}

func (wt seqWithTest) WithBytes(p []byte) Sequence {
	return seqWithTest{wt.info, wt.feats, p}
}

func TestSeqWithInterface(t *testing.T) {
	p := []byte(strings.Repeat("atgc", 100))
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{NewFeature("source", Range(0, len(p)), props)}
	info := "info"

	in := newSeqWithTest(nil, nil, nil)
	out := WithInfo(in, info)
	testutils.Equals(t, out, newSeqWithTest(info, nil, nil))

	out = WithFeatures(in, ff)
	testutils.Equals(t, out, newSeqWithTest(nil, ff, nil))

	out = WithBytes(in, p)
	testutils.Equals(t, out, newSeqWithTest(nil, nil, p))
}

var searchTests = []struct {
	in  Sequence
	out []Segment
}{
	{New(nil, nil, []byte("atgc")), []Segment{{0, 4}, {4, 8}, {8, 12}}},
	{New(nil, nil, []byte("")), nil},
}

func TestSearch(t *testing.T) {
	seq := New(nil, nil, []byte("atgcatgcatgc"))
	for _, tt := range searchTests {
		out := Search(seq, tt.in)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("Search(%v, %v) = %v, want %v", seq, tt.in, out, tt.out)
		}
	}
}
