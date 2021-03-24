package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func TestComplement(t *testing.T) {
	p, q := []byte("ACGTURYKMWSBDHVacgturykmwsbdhv."), []byte("TGCAAYRMKSWVHDBtgcaayrmkswvhdb.")
	props := Props{}
	props.Add("organism", "Genus species")
	props.Add("mol_type", "Genomic DNA")
	ff := []Feature{
		NewFeature("source", Range(0, len(p)), props),
		NewFeature("gene", Range(2, 4), props),
		NewFeature("misc_feature", Ambiguous{5, 7}, props),
	}
	gg := []Feature{
		NewFeature("source", Range(0, len(p)).Complement(), props),
		NewFeature("gene", Range(2, 4).Complement(), props),
		NewFeature("misc_feature", Ambiguous{5, 7}, props),
	}
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

var matchTests = []struct {
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

func TestMatch(t *testing.T) {
	for _, tt := range matchTests {
		query := New(nil, nil, []byte{tt.base})
		seq := New(nil, nil, []byte(tt.match))
		exp := make([]Segment, len(tt.match))
		for i := range exp {
			exp[i] = Segment{i, i + 1}
		}
		out := Match(seq, query)
		if len(out) != len(exp) {
			testutils.Equals(t, out, exp)
		}
	}
}
