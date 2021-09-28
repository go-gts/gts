package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func TestComplement(t *testing.T) {
	in := AsSequence("ACGTURYKMSWBDHVacgturykmswbdhv.")
	exp := AsSequence("TGCAAYRMKSWVHDBtgcaayrmkswvhdb.")
	out := Complement(in)
	testutils.Diff(t, string(out.Bytes()), string(exp.Bytes()))
}

func TestTranscribe(t *testing.T) {
	in := AsSequence("ACGTURYKMSWBDHVacgturykmswbdhv.")
	exp := AsSequence("UGCAAYRMKSWVHDBugcaayrmkswvhdb.")
	out := Transcribe(in)
	testutils.Diff(t, string(out.Bytes()), string(exp.Bytes()))
}

var matchTests = []struct {
	base  byte
	match string
}{
	{'a', ""},                 // case 1
	{'a', "a"},                // case 2
	{'c', "c"},                // case 3
	{'g', "g"},                // case 4
	{'t', "tu"},               // case 5
	{'u', "tu"},               // case 6
	{'r', "agr"},              // case 7
	{'y', "ctuy"},             // case 8
	{'k', "gtuy"},             // case 9
	{'m', "acm"},              // case 10
	{'s', "cgs"},              // case 11
	{'w', "atuw"},             // case 12
	{'b', "cgtuyksb"},         // case 13
	{'d', "agturkwd"},         // case 14
	{'h', "actuymwh"},         // case 15
	{'v', "acgrmsv"},          // case 16
	{'n', "acgturykmswbdhvn"}, // case 17
}

func TestMatch(t *testing.T) {
	for i, tt := range matchTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			query := AsSequence(tt.base)
			seq := AsSequence(tt.match)
			exp := make([]Segment, len(tt.match))
			for i := range exp {
				exp[i] = Segment{i, i + 1}
			}
			out := Match(seq, query)
			if len(out) != len(exp) {
				testutils.Equals(t, out, exp)
			}
		})
	}
}
