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
	{"1", locationLocator(Point(0))},
	{"3..6", locationLocator(Range(2, 6))},
	{"complement(3..6)", locationLocator(Range(2, 6).Complement())},

	{"exon", filterLocator(selectorFilter("exon"))},
	{"exon/gene=INS", filterLocator(selectorFilter("exon"))},
	{"/gene=INS", filterLocator(selectorFilter("/gene=INS"))},

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
	result, err := FeatureTableParser("").Parse(pars.FromString(featureIOTest))
	if err != nil {
		t.Errorf("failed to parse feature table: %v", err)
		return
	}
	ff := result.Value.(FeatureTable)
	for _, tt := range asLocatorTests {
		loc, err := AsLocator(tt.in)
		if err != nil {
			t.Errorf("AsLocator(%q): %v", tt.in, err)
			return
		}
		out := loc(ff)
		exp := tt.loc(ff)
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
