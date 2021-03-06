package gts

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-pars/pars"
	"github.com/go-test/deep"
)

var featureIOTests = []string{
	`     source          1..465
                     /organism="Homo sapiens"
                     /mol_type="mRNA"
                     /db_xref="taxon:9606"
                     /chromosome="11"
                     /map="11p15.5"
     gene            1..465
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /note="insulin"
                     /db_xref="GeneID:3630"
                     /db_xref="HGNC:HGNC:6081"
                     /db_xref="MIM:176730"
     exon            1..42
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /inference="alignment:Splign:2.1.0"
     exon            43..246
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /inference="alignment:Splign:2.1.0"
     CDS             60..392
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /note="proinsulin; preproinsulin"
                     /codon_start=1
                     /product="insulin preproprotein"
                     /protein_id="NP_000198.1"
                     /db_xref="CCDS:CCDS7729.1"
                     /db_xref="GeneID:3630"
                     /db_xref="HGNC:HGNC:6081"
                     /db_xref="MIM:176730"
                     /translation="MALWMRLLPLLALLALWGPDPAAAFVNQHLCGSHLVEALYLVCG
                     ERGFFYTPKTRREAEDLQVGQVELGGGPGAGSLQPLALEGSLQKRGIVEQCCTSICSL
                     YQLENYCN"
     sig_peptide     60..131
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /inference="COORDINATES: ab initio prediction:SignalP:4.0"
     proprotein      132..389
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /product="proinsulin"
     mat_peptide     132..221
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /product="insulin B chain"
     mat_peptide     228..320
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /product="C-peptide"
     mat_peptide     327..389
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /product="insulin A chain"
     exon            247..465
                     /gene="INS"
                     /gene_synonym="IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"
                     /inference="alignment:Splign:2.1.0"`,
	`     tRNA            complement(3838377..3838450)
                     /gene="TRI-GAT1-1"
                     /product="tRNA-Ile"
                     /inference="COORDINATES: profile:tRNAscan-SE:1.23"
                     /note="tRNA-Ile (anticodon GAT) 1-1; Derived by automated
                     computational analysis using gene prediction method:
                     tRNAscan-SE."
                     /anticodon=(pos:complement(3838414..3838416),aa:Ile,
                     seq:gat)
                     /db_xref="GeneID:100189132"
                     /db_xref="HGNC:HGNC:34694"`,
}

func TestFeatureKeylineParser(t *testing.T) {
	parser := pars.Exact(featureKeylineParser("     ", 21))
	for _, in := range []string{
		"     source          ",
		"    source          ",
		"     ",
		"     source",
		"     source 1..39",
		"     source          1..39 ",
	} {
		state := pars.FromString(in)
		if _, err := parser.Parse(state); err == nil {
			t.Errorf("while parsing`\n%s\n`: expected error", in)
		}
	}
}

func TestFeatureIO(t *testing.T) {
	parser := pars.Exact(FeatureTableParser(""))
	for _, in := range featureIOTests {
		state := pars.FromString(in)
		result, err := parser.Parse(state)
		if err != nil {
			t.Errorf("while parsing`\n%s\n`: %v", in, err)
			return
		}
		switch ff := result.Value.(type) {
		case FeatureTable:
			b := strings.Builder{}
			n, err := ff.Format("     ", 21).WriteTo(&b)
			if err != nil {
				t.Errorf("f.WriteTo(w) = %d, %v, want %d, nil", n, err, n)
			}
			out := b.String()
			if out != in {
				t.Errorf("f.Format(%q, 21) = %q, want %q", "     ", out, in)
			}
		default:
			t.Errorf("result.Value.(type) = %T, want %T", ff, FeatureTable{})
		}
	}

	if err := parser(pars.FromString(""), pars.Void); err == nil {
		t.Error("while parsing empty string: expected error")
	}
}

func TestFeature(t *testing.T) {
	key := "gene"
	loc := Range(1, 465)
	qfs := Values{}
	qfs.Set("gene", "INS")
	qfs.Set("db_xref", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730")
	f := Feature{key, loc, qfs, map[string]int{"gene": 0, "note": 1}}
	qq := listQualifiers(f)
	out := []QualifierIO{
		{"gene", "INS"},
		{"db_xref", "GeneID:3630"},
		{"db_xref", "HGNC:HGNC:6081"},
		{"db_xref", "MIM:176730"},
	}
	testutils.Equals(t, qq, out)
}

func TestFeatureRepair(t *testing.T) {
	state := pars.FromString(featureIOTests[0])
	parser := pars.Exact(FeatureTableParser(""))
	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("while parsing`\n%s\n`: %v", featureIOTests[0], err)
		return
	}
	m, n := 200, 465
	exp := result.Value.(FeatureTable)
	seq := Sequence(New(nil, exp, bytes.Repeat([]byte("a"), n)))
	left, right := Slice(seq, 0, m), Slice(seq, m, n)
	seq = Concat(left, right)

	in := seq.Features()
	out := FeatureTable(Repair(in))

	sort.Sort(out)
	sort.Sort(exp)

	if !featuresEqual(out, exp) {
		sin := in.Format("     ", 21).String()
		sout := out.Format("     ", 21).String()
		sexp := exp.Format("     ", 21).String()
		t.Errorf("Repair: \n%s\nDiff:", sin)
		testutils.DiffLine(t, sout, sexp)
	}
}

var sampleSourceFeature = Feature{"source", Range(0, 5386), Values{"mol_type": []string{"Genomic DNA"}}, nil}
var sampleGeneFeature = Feature{"gene", Range(51, 221), Values{"locus_tag": []string{"phiX174p04"}}, nil}
var sampleCDSFeature = Feature{"CDS", Range(133, 393), Values{"locus_tag": []string{"phiX174p05"}}, nil}
var sampleFeatureTable = FeatureTable{
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
	out FeatureTable
}{
	{TrueFilter, sampleFeatureTable},
	{FalseFilter, FeatureTable{}},
	{Key(""), sampleFeatureTable},
	{Key("source"), FeatureTable{sampleSourceFeature}},
	{Key("gene"), FeatureTable{sampleGeneFeature}},
	{qualifierFilter("mol_type", "DNA"), FeatureTable{sampleSourceFeature}},
	{qualifierFilter("", "DNA"), FeatureTable{sampleSourceFeature}},
	{And(Key("source"), Key("gene")), FeatureTable{}},
	{And(Key("source"), qualifierFilter("mol_type", "DNA")), FeatureTable{sampleSourceFeature}},
	{Or(Key("source"), Key("gene")), sampleFeatureTable},
	{Or(Key("foo"), Key("bar")), FeatureTable{}},
	{Not(Key("source")), FeatureTable{sampleGeneFeature}},
	{selectorFilter("source/mol_type=DNA"), FeatureTable{sampleSourceFeature}},
	{selectorFilter("source/mol_type"), FeatureTable{sampleSourceFeature}},
	{selectorFilter("source/mol_type=\\/"), FeatureTable{}},
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

func TestFeatureTableSort(t *testing.T) {
	state := pars.FromString(featureIOTests[0])
	parser := pars.Exact(FeatureTableParser(""))
	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("while parsing`\n%s\n`: %v", featureIOTests[0], err)
		return
	}

	in := result.Value.(FeatureTable)
	out := make(FeatureTable, len(in))
	exp := make(FeatureTable, len(in))

	sort.Sort(in)
	copy(exp, in)
	sort.Sort(sort.Reverse(in))
	copy(out, in)
	sort.Sort(out)

	if !featuresEqual(out, exp) {
		sin := in.Format("     ", 21).String()
		sout := out.Format("     ", 21).String()
		sexp := exp.Format("     ", 21).String()
		t.Errorf("Repair: \n%s\nDiff:", sin)
		testutils.DiffLine(t, sout, sexp)
	}
}

func TestFeatureInsert(t *testing.T) {
	ff := FeatureTable{}
	ff = ff.Insert(sampleCDSFeature)
	testutils.Equals(t, ff, FeatureTable{sampleCDSFeature})
	ff = ff.Insert(sampleSourceFeature)
	testutils.Equals(t, ff, FeatureTable{sampleSourceFeature, sampleCDSFeature})
	ff = ff.Insert(sampleGeneFeature)
	testutils.Equals(t, ff, FeatureTable{sampleSourceFeature, sampleGeneFeature, sampleCDSFeature})
}
