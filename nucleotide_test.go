package gts

import (
	"testing"

	"github.com/go-gts/gts/testutils"
)

func TestComplement(t *testing.T) {
	p, q := []byte("ACGTURYKMBDHVacgturykmbdhv."), []byte("TGCAAYRMKVHDBtgcaayrmkvhdb.")
	qfs := Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	ff := []Feature{{"source", Range(0, len(p)), qfs, nil}, {"gene", Range(2, 4), qfs, nil}, {"misc_feature", Ambiguous{5, 7}, qfs, nil}}
	gg := []Feature{{"source", Range(0, len(p)).Complement(), qfs, nil}, {"gene", Range(2, 4).Complement(), qfs, nil}, {"misc_feature", Ambiguous{5, 7}, qfs, nil}}
	in := New(nil, ff, p)
	exp := New(nil, gg, q)
	out := Complement(in)
	testutils.Equals(t, out, exp)
}

func TestTranscribe(t *testing.T) {
	in := New(nil, nil, []byte("ACGTURYKMBDHVacgturykmbdhv."))
	exp := New(nil, nil, []byte("UGCAAYRMKVHDBtgcaayrmkvhdb."))
	out := Transcribe(in)
	testutils.Equals(t, out, exp)
}
