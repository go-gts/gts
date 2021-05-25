package gts

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-test/deep"
)

var testFeatureTable = []Feature{
	NewFeature("source", Range(0, 465), Props{
		[]string{"chromosome", "11"},
		[]string{"db_xref", "taxon:9606"},
		[]string{"map", "11p15.5"},
		[]string{"mol_type", "mRNA"},
		[]string{"organism", "Homo sapiens"},
	}),
	NewFeature("gene", Range(0, 465), Props{
		[]string{"db_xref", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"note", "insulin"},
	}),
	NewFeature("exon", Range(0, 42), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"inference", "alignment:Splign:2.1.0"},
	}),
	NewFeature("exon", Range(42, 246), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"inference", "alignment:Splign:2.1.0"},
	}),
	NewFeature("CDS", Range(59, 392), Props{
		[]string{"codon_start", "1"},
		[]string{"db_xref", "CCDS:CCDS7729.1", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"note", "proinsulin; preproinsulin"},
		[]string{"product", "insulin preproprotein"},
		[]string{"protein_id", "NP_000198.1"},
		[]string{"translation", "MALWMRLLPLLALLALWGPDPAAAFVNQHLCGSHLVEALYLVCG\nERGFFYTPKTRREAEDLQVGQVELGGGPGAGSLQPLALEGSLQKRGIVEQCCTSICSL\nYQLENYCN"},
	}),
	NewFeature("sig_peptide", Range(59, 131), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"inference", "COORDINATES: ab initio prediction:SignalP:4.0"},
	}),
	NewFeature("proprotein", Range(131, 389), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"product", "proinsulin"},
	}),
	NewFeature("mat_peptide", Range(131, 221), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"product", "insulin B chain"},
	}),
	NewFeature("mat_peptide", Range(227, 320), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"product", "C-peptide"},
	}),
	NewFeature("mat_peptide", Range(326, 389), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"product", "insulin A chain"},
	}),
	NewFeature("exon", Range(246, 465), Props{
		[]string{"gene", "INS"},
		[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
		[]string{"inference", "alignment:Splign:2.1.0"},
	}),
}

func TestFeatureRepair(t *testing.T) {
	m, n := 200, 465
	exp := testFeatureTable
	seq := Sequence(New(nil, exp, bytes.Repeat([]byte("a"), n)))
	left, right := Slice(seq, 0, m), Slice(seq, m, n)
	seq = Concat(left, right)

	in := seq.Features()
	out := Repair(in)

	sort.Sort(FeatureSlice(out))
	sort.Sort(FeatureSlice(exp))

	if !featuresEqual(out, exp) {
		ssin := make([]string, len(in))
		for i, f := range in {
			ssin[i] = jsonify(f)
		}

		ssout := make([]string, len(out))
		for i, f := range out {
			ssout[i] = jsonify(f)
		}

		ssexp := make([]string, len(exp))
		for i, f := range exp {
			ssexp[i] = jsonify(f)
		}

		sin := strings.Join(ssin, "\n")
		sout := strings.Join(ssout, "\n")
		sexp := strings.Join(ssexp, "\n")

		t.Errorf("Repair: \n%s\nDiff:", sin)
		testutils.DiffLine(t, sout, sexp)
	}
}

var sampleSourceFeature = NewFeature("source", Range(0, 5386), Props{[]string{"mol_type", "Genomic DNA"}})
var sampleGeneFeature = NewFeature("gene", Range(51, 221), Props{[]string{"locus_tag", "phiX174p04"}})
var sampleCDSFeature = NewFeature("CDS", Range(133, 393), Props{[]string{"locus_tag", "phiX174p05"}})
var sampleFeatureTable = FeatureSlice{
	sampleSourceFeature,
	sampleGeneFeature,
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

var featureFilterTests = []struct {
	f   Filter
	out FeatureSlice
}{
	{TrueFilter, sampleFeatureTable},
	{FalseFilter, FeatureSlice{}},
	{Key(""), sampleFeatureTable},
	{Key("source"), FeatureSlice{sampleSourceFeature}},
	{Key("gene"), FeatureSlice{sampleGeneFeature}},
	{qualifierFilter("mol_type", "DNA"), FeatureSlice{sampleSourceFeature}},
	{qualifierFilter("", "DNA"), FeatureSlice{sampleSourceFeature}},
	{And(), sampleFeatureTable},
	{And(Key("source"), Key("gene")), FeatureSlice{}},
	{And(Key("source"), qualifierFilter("mol_type", "DNA")), FeatureSlice{sampleSourceFeature}},
	{Or(), sampleFeatureTable},
	{Or(Key("source"), Key("gene")), sampleFeatureTable},
	{Or(Key("foo"), Key("bar")), FeatureSlice{}},
	{Not(Key("source")), FeatureSlice{sampleGeneFeature}},
	{selectorFilter("source/mol_type=DNA"), FeatureSlice{sampleSourceFeature}},
	{selectorFilter("source/mol_type"), FeatureSlice{sampleSourceFeature}},
	{selectorFilter("source/mol_type=\\/"), FeatureSlice{}},
	{ForwardStrand, sampleFeatureTable},
	{ReverseStrand, FeatureSlice{}},
}

func TestFeatureFilter(t *testing.T) {
	for i, tt := range featureFilterTests {
		out := sampleFeatureTable.Filter(tt.f)
		if diff := deep.Equal(out, tt.out); diff != nil {
			t.Errorf("case %d: %v", i+1, diff)
		}
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
func TestFeatureSliceSort(t *testing.T) {
	in := testFeatureTable
	out := make([]Feature, len(in))
	exp := make([]Feature, len(in))

	sort.Sort(FeatureSlice(in))
	copy(exp, in)
	sort.Sort(sort.Reverse(FeatureSlice(in)))
	copy(out, in)
	sort.Sort(FeatureSlice(out))

	if !featuresEqual(out, exp) {
		t.Error("sorted slices differ")
	}
}

func TestFeatureInsert(t *testing.T) {
	ff := FeatureSlice{}
	ff = ff.Insert(sampleCDSFeature)
	testutils.Equals(t, ff, FeatureSlice{sampleCDSFeature})
	ff = ff.Insert(sampleSourceFeature)
	testutils.Equals(t, ff, FeatureSlice{sampleSourceFeature, sampleCDSFeature})
	ff = ff.Insert(sampleGeneFeature)
	testutils.Equals(t, ff, FeatureSlice{sampleSourceFeature, sampleGeneFeature, sampleCDSFeature})
}
