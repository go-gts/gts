package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-test/deep"
)

var asLocatorTests = []struct {
	in  string
	loc Locator
}{
	{"^..$", relativeLocator(HeadTail{0, 0})},                       // case 1
	{"1", locationLocator(Point(0))},                                // case 2
	{"3..6", locationLocator(Range(2, 6))},                          // case 3
	{"complement(3..6)", locationLocator(Range(2, 6).Complement())}, // case 4

	{"exon", filterLocator(selectorFilter("exon"))},           // case 5
	{"exon/gene=INS", filterLocator(selectorFilter("exon"))},  // case 6
	{"/gene=INS", filterLocator(selectorFilter("/gene=INS"))}, // case 7

	{"@^-20..^", resizeLocator(allLocator, HeadHead{-20, 0})},                           // case 8
	{"@^..$", resizeLocator(allLocator, HeadTail{0, 0})},                                // case 9
	{"exon@^..$", resizeLocator(filterLocator(selectorFilter("exon")), HeadTail{0, 0})}, // case 10
}

var asLocatorFailTests = []string{
	"exon/gene=[",    // case 1
	"@",              // case 2
	"exon/gene=[@",   // case 3
	"exon/gene=INS@", // case 4
}

func TestAsLocator(t *testing.T) {
	ff := testFeatureTable
	seq := Concat(
		AsSequence("AGCCCTCCAGGACAGGCTGCATCAGAAGAGGCCATCAAGCAGATCACTGTCCTTCTGCCATGGCCCTGTG"),
		AsSequence("GATGCGCCTCCTGCCCCTGCTGGCGCTGCTGGCCCTCTGGGGACCTGACCCAGCCGCAGCCTTTGTGAAC"),
		AsSequence("CAACACCTGTGCGGCTCACACCTGGTGGAAGCTCTCTACCTAGTGTGCGGGGAACGAGGCTTCTTCTACA"),
		AsSequence("CACCCAAGACCCGCCGGGAGGCAGAGGACCTGCAGGTGGGGCAGGTGGAGCTGGGCGGGGGCCCTGGTGC"),
		AsSequence("AGGCAGCCTGCAGCCCTTGGCCCTGGAGGGGTCCCTGCAGAAGCGTGGCATTGTGGAACAATGCTGTACC"),
		AsSequence("AGCATCTGCTCCCTCTACCAGCTGGAGAACTACTGCAACTAGACGCAGCCCGCAGGCAGCCCCACACCCG"),
		AsSequence("CCGCCTCCTGCACCGAGAGAGATGGAATAAAGCCCTTGAACCAGC"),
	)

	for i, tt := range asLocatorTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			loc, err := AsLocator(tt.in)
			if err != nil {
				t.Errorf("AsLocator(%q): %v", tt.in, err)
				return
			}
			out := loc(ff, seq)
			exp := tt.loc(ff, seq)
			if diff := deep.Equal(out, exp); diff != nil {
				t.Errorf("AsLocator(%q): %v", tt.in, diff)
			}
		})
	}

	for i, in := range asLocatorFailTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			_, err := AsLocator(in)
			if err == nil {
				t.Errorf("AsLocator(%q) expected an error", in)
			}
		})
	}
}
