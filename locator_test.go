package gts

import (
	"testing"

	"github.com/go-pars/pars"
	"github.com/go-test/deep"
)

var asLocatorTests = []struct {
	in  string
	loc Locator
}{
	{"^..$", relativeLocator(HeadTail{0, 0})},
	{"1", locationLocator(Point(0))},
	{"3..6", locationLocator(Range(2, 6))},
	{"complement(3..6)", locationLocator(Range(2, 6).Complement())},

	{"exon", filterLocator(selectorFilter("exon"))},
	{"exon/gene=INS", filterLocator(selectorFilter("exon"))},
	{"/gene=INS", filterLocator(selectorFilter("/gene=INS"))},

	{"@^-20..^", resizeLocator(allLocator, HeadHead{-20, 0})},
	{"@^..$", resizeLocator(allLocator, HeadTail{0, 0})},
	{"exon@^..$", resizeLocator(filterLocator(selectorFilter("exon")), HeadTail{0, 0})},
}

var asLocatorFailTests = []string{
	"exon/gene=[",
	"@",
	"exon/gene=[@",
	"exon/gene=INS@",
}

func TestAsLocator(t *testing.T) {
	result, err := FeatureTableParser("").Parse(pars.FromString(featureIOTests[0]))
	if err != nil {
		t.Errorf("failed to parse feature table: %v", err)
		return
	}
	ff := result.Value.(FeatureTable)
	seq := New(nil, ff, []byte(""+
		"AGCCCTCCAGGACAGGCTGCATCAGAAGAGGCCATCAAGCAGATCACTGTCCTTCTGCCATGGCCCTGTG"+
		"GATGCGCCTCCTGCCCCTGCTGGCGCTGCTGGCCCTCTGGGGACCTGACCCAGCCGCAGCCTTTGTGAAC"+
		"CAACACCTGTGCGGCTCACACCTGGTGGAAGCTCTCTACCTAGTGTGCGGGGAACGAGGCTTCTTCTACA"+
		"CACCCAAGACCCGCCGGGAGGCAGAGGACCTGCAGGTGGGGCAGGTGGAGCTGGGCGGGGGCCCTGGTGC"+
		"AGGCAGCCTGCAGCCCTTGGCCCTGGAGGGGTCCCTGCAGAAGCGTGGCATTGTGGAACAATGCTGTACC"+
		"AGCATCTGCTCCCTCTACCAGCTGGAGAACTACTGCAACTAGACGCAGCCCGCAGGCAGCCCCACACCCG"+
		"CCGCCTCCTGCACCGAGAGAGATGGAATAAAGCCCTTGAACCAGC"))
	for _, tt := range asLocatorTests {
		loc, err := AsLocator(tt.in)
		if err != nil {
			t.Errorf("AsLocator(%q): %v", tt.in, err)
			return
		}
		out := loc(seq)
		exp := tt.loc(seq)
		if diff := deep.Equal(out, exp); diff != nil {
			t.Errorf("AsLocator(%q): %v", tt.in, diff)
		}
	}

	for _, in := range asLocatorFailTests {
		_, err := AsLocator(in)
		if err == nil {
			t.Errorf("AsLocator(%q) expected an error", in)
		}
	}
}
