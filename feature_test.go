package gts

import (
	"sort"
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-gts/gts/perm"
)

var sampleSourceFeature = NewFeature("source", Range(0, 5386), Props{[]string{"mol_type", "Genomic DNA"}})
var sampleGeneFeature = NewFeature("gene", Range(51, 221), Props{[]string{"locus_tag", "phiX174p04"}})
var sampleCDSFeature = NewFeature("CDS", Range(133, 393), Props{[]string{"locus_tag", "phiX174p05"}})
var sampleFeatureTable = Features{
	sampleSourceFeature,
	sampleGeneFeature,
	sampleCDSFeature,
}

func qualifierFilter(name, exp string) Filter {
	f, err := Qualifier(name, exp)
	if err != nil {
		panic(err)
	}
	return f
}

func selectorFilter(sel string) Filter {
	f, err := Selector(sel)
	if err != nil {
		panic(err)
	}
	return f
}

func serializeFeatures(ff Features) string {
	ss := make([]string, len(ff))
	for i, f := range ff {
		ss[i] = jsonify(f)
	}
	return strings.Join(ss, "\n")
}

func featuresDiff(t *testing.T, a, b Features) {
	testutils.DiffLine(t, serializeFeatures(a), serializeFeatures(b))
}

var featureFilterTests = []struct {
	f   Filter
	out Features
}{
	{TrueFilter, sampleFeatureTable},
	{FalseFilter, Features{}},
	{Within(50, 250), Features{sampleGeneFeature}},
	{Overlap(300, 400), Features{sampleSourceFeature, sampleCDSFeature}},
	{Key(""), sampleFeatureTable},
	{Key("source"), Features{sampleSourceFeature}},
	{Key("gene"), Features{sampleGeneFeature}},
	{qualifierFilter("mol_type", "DNA"), Features{sampleSourceFeature}},
	{qualifierFilter("", "DNA"), Features{sampleSourceFeature}},
	{And(), sampleFeatureTable},
	{And(Key("source"), Key("gene")), Features{}},
	{And(Key("source"), qualifierFilter("mol_type", "DNA")), Features{sampleSourceFeature}},
	{Or(), sampleFeatureTable},
	{Or(Key("source"), Key("gene")), Features{sampleSourceFeature, sampleGeneFeature}},
	{Or(Key("foo"), Key("bar")), Features{}},
	{Not(Key("source")), Features{sampleGeneFeature, sampleCDSFeature}},
	{selectorFilter("source/mol_type=DNA"), Features{sampleSourceFeature}},
	{selectorFilter("source/mol_type"), Features{sampleSourceFeature}},
	{selectorFilter("source/mol_type=\\/"), Features{}},
	{ForwardStrand, sampleFeatureTable},
	{ReverseStrand, Features{}},
}

func TestFeatureFilter(t *testing.T) {
	for i, tt := range featureFilterTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := sampleFeatureTable.Filter(tt.f)
			testutils.DiffLine(t, serializeFeatures(out), serializeFeatures(tt.out))
		})
	}
}

func TestFeatureQualifierFilter(t *testing.T) {
	testutils.Panics(t, func() {
		qualifierFilter("", "[")
	})

	testutils.Panics(t, func() {
		selectorFilter("/mol_type=[")
	})
}

func TestFeaturesSort(t *testing.T) {
	ff := make(Features, len(sampleFeatureTable))
	copy(ff, sampleFeatureTable)

	iter := perm.Permutate(ff)
	for i := 0; iter.Next(); i++ {
		testutils.RunCase(t, i, func(t *testing.T) {
			sort.Sort(ff)
			featuresDiff(t, ff, sampleFeatureTable)
		})
	}
}

func TestFeatureInsert(t *testing.T) {
	ff := Features{}

	ff = ff.Insert(sampleCDSFeature)
	featuresDiff(t, ff, Features{sampleCDSFeature})

	ff = ff.Insert(sampleSourceFeature)
	featuresDiff(t, ff, Features{sampleSourceFeature, sampleCDSFeature})

	ff = ff.Insert(sampleGeneFeature)
	featuresDiff(t, ff, Features{sampleSourceFeature, sampleGeneFeature, sampleCDSFeature})
}
