package gts

import (
	"testing"

	"github.com/go-gts/gts/testutils"
)

func TestComplement(t *testing.T) {
	p, q := []byte("ACGTURYKMWSBDHVacgturykmwsbdhv."), []byte("TGCAAYRMKSWVHDBtgcaayrmkswvhdb.")
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
	in := New(nil, nil, []byte("ACGTURYKMWSBDHVacgturykmwsbdhv."))
	exp := New(nil, nil, []byte("UGCAAYRMKSWVHDBtgcaayrmkswvhdb."))
	out := Transcribe(in)
	testutils.Equals(t, out, exp)
}

var searchTests = []struct {
	base  byte
	match string
}{
	{'a', ""},
	{'a', "a"},
	{'c', "c"},
	{'g', "g"},
	{'t', "tu"},
	{'u', "tu"},
	{'r', "agr"},
	{'y', "ctuy"},
	{'k', "gtuy"},
	{'m', "acm"},
	{'s', "cgs"},
	{'w', "atuw"},
	{'b', "cgtuyksb"},
	{'d', "agturkwd"},
	{'h', "actuymwh"},
	{'v', "acgrmsv"},
	{'n', "acgturykmswbdhvn"},
}

func TestSearch(t *testing.T) {
	for _, tt := range searchTests {
		query := New(nil, nil, []byte{tt.base})
		seq := New(nil, nil, []byte(tt.match))
		exp := make([]Location, len(tt.match))
		for i := range exp {
			exp[i] = Range(i, i+1)
		}
		out := Search(seq, query)
		if len(out) != len(exp) {
			testutils.Equals(t, out, exp)
		}
	}
}
