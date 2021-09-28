package gts

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

var regionAccessorTests = []struct {
	in   Region
	len  int
	head int
	tail int
}{
	{Segment{3, 6}, 3, 3, 6},
	{Segment{6, 3}, 3, 6, 3},
	{Regions{}, 0, 0, 0},
	{Regions{Segment{3, 6}, Segment{13, 16}}, 6, 3, 16},
}

func TestRegionAccessor(t *testing.T) {
	for i, tt := range regionAccessorTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			if tt.in.Len() != tt.len {
				t.Errorf("%#v.Len() = %d, want %d", tt.in, tt.in.Len(), tt.len)
			}

			if tt.in.Head() != tt.head {
				t.Errorf("%#v.Head() = %d, want %d", tt.in, tt.in.Head(), tt.head)
			}

			if tt.in.Tail() != tt.tail {
				t.Errorf("%#v.Tail() = %d, want %d", tt.in, tt.in.Tail(), tt.len)
			}
		})
	}
}

var regionResizeTests = []struct {
	in       Region
	modifier Modifier
	out      Region
}{
	{Segment{3, 6}, Head(+0), Segment{3, 3}}, // case 1
	{Segment{3, 6}, Head(+1), Segment{4, 4}}, // case 2
	{Segment{3, 6}, Head(-1), Segment{2, 2}}, // case 3

	{Segment{3, 6}, Tail(+0), Segment{6, 6}}, // case 4
	{Segment{3, 6}, Tail(+1), Segment{7, 7}}, // case 5
	{Segment{3, 6}, Tail(-1), Segment{5, 5}}, // case 6

	{Segment{3, 6}, HeadTail{+0, +0}, Segment{3, 6}}, // case 7
	{Segment{3, 6}, HeadTail{+0, +1}, Segment{3, 7}}, // case 8
	{Segment{3, 6}, HeadTail{+2, +0}, Segment{5, 6}}, // case 9
	{Segment{3, 6}, HeadTail{+0, -1}, Segment{3, 5}}, // case 10
	{Segment{3, 6}, HeadTail{-2, +0}, Segment{1, 6}}, // case 11
	{Segment{3, 6}, HeadTail{+2, -1}, Segment{5, 5}}, // case 12
	{Segment{3, 6}, HeadTail{+2, -2}, Segment{5, 5}}, // case 13
	{Segment{3, 6}, HeadTail{-2, +1}, Segment{1, 7}}, // case 14
	{Segment{3, 6}, HeadTail{+2, +1}, Segment{5, 7}}, // case 15
	{Segment{3, 6}, HeadTail{-2, -1}, Segment{1, 5}}, // case 16

	{Segment{3, 6}, HeadHead{+0, +0}, Segment{3, 3}}, // case 17
	{Segment{3, 6}, HeadHead{+0, +1}, Segment{3, 4}}, // case 18
	{Segment{3, 6}, HeadHead{+2, +0}, Segment{5, 5}}, // case 19
	{Segment{3, 6}, HeadHead{+0, -1}, Segment{3, 3}}, // case 20
	{Segment{3, 6}, HeadHead{-2, +0}, Segment{1, 3}}, // case 21
	{Segment{3, 6}, HeadHead{+2, -1}, Segment{5, 5}}, // case 22
	{Segment{3, 6}, HeadHead{-2, +1}, Segment{1, 4}}, // case 23
	{Segment{3, 6}, HeadHead{+2, +1}, Segment{5, 5}}, // case 24
	{Segment{3, 6}, HeadHead{-2, -1}, Segment{1, 2}}, // case 25

	{Segment{3, 6}, TailTail{+0, +0}, Segment{6, 6}}, // case 26
	{Segment{3, 6}, TailTail{+0, +1}, Segment{6, 7}}, // case 27
	{Segment{3, 6}, TailTail{+2, +0}, Segment{8, 8}}, // case 28
	{Segment{3, 6}, TailTail{+0, -1}, Segment{6, 6}}, // case 29
	{Segment{3, 6}, TailTail{-2, +0}, Segment{4, 6}}, // case 30
	{Segment{3, 6}, TailTail{+2, -1}, Segment{8, 8}}, // case 31
	{Segment{3, 6}, TailTail{-2, +1}, Segment{4, 7}}, // case 32
	{Segment{3, 6}, TailTail{+2, +1}, Segment{8, 8}}, // case 33
	{Segment{3, 6}, TailTail{-2, -1}, Segment{4, 5}}, // case 34

	{Segment{6, 3}, Head(+0), Segment{6, 6}}, // case 35
	{Segment{6, 3}, Head(+1), Segment{5, 5}}, // case 36
	{Segment{6, 3}, Head(-1), Segment{7, 7}}, // case 37

	{Segment{6, 3}, Tail(+0), Segment{3, 3}}, // case 38
	{Segment{6, 3}, Tail(+1), Segment{2, 2}}, // case 39
	{Segment{6, 3}, Tail(-1), Segment{4, 4}}, // case 40

	{Segment{6, 3}, HeadTail{+0, +0}, Segment{6, 3}}, // case 41
	{Segment{6, 3}, HeadTail{+0, +1}, Segment{6, 2}}, // case 42
	{Segment{6, 3}, HeadTail{+2, +0}, Segment{4, 3}}, // case 43
	{Segment{6, 3}, HeadTail{+0, -1}, Segment{6, 4}}, // case 44
	{Segment{6, 3}, HeadTail{-2, +0}, Segment{8, 3}}, // case 45
	{Segment{6, 3}, HeadTail{+2, -1}, Segment{4, 4}}, // case 46
	{Segment{6, 3}, HeadTail{-2, +1}, Segment{8, 2}}, // case 47
	{Segment{6, 3}, HeadTail{+2, +1}, Segment{4, 2}}, // case 48
	{Segment{6, 3}, HeadTail{-2, -1}, Segment{8, 4}}, // case 49

	{Segment{6, 3}, HeadHead{+0, +0}, Segment{6, 6}}, // case 50
	{Segment{6, 3}, HeadHead{+0, +1}, Segment{6, 5}}, // case 51
	{Segment{6, 3}, HeadHead{+2, +0}, Segment{4, 4}}, // case 52
	{Segment{6, 3}, HeadHead{+0, -1}, Segment{6, 6}}, // case 53
	{Segment{6, 3}, HeadHead{-2, +0}, Segment{8, 6}}, // case 54
	{Segment{6, 3}, HeadHead{+2, -1}, Segment{4, 4}}, // case 55
	{Segment{6, 3}, HeadHead{-2, +1}, Segment{8, 5}}, // case 56
	{Segment{6, 3}, HeadHead{+2, +1}, Segment{4, 4}}, // case 57
	{Segment{6, 3}, HeadHead{-2, -1}, Segment{8, 7}}, // case 58

	{Segment{6, 3}, TailTail{+0, +0}, Segment{3, 3}}, // case 59
	{Segment{6, 3}, TailTail{+0, +1}, Segment{3, 2}}, // case 60
	{Segment{6, 3}, TailTail{+2, +0}, Segment{1, 1}}, // case 61
	{Segment{6, 3}, TailTail{+0, -1}, Segment{3, 3}}, // case 62
	{Segment{6, 3}, TailTail{-2, +0}, Segment{5, 3}}, // case 63
	{Segment{6, 3}, TailTail{+2, -1}, Segment{1, 1}}, // case 64
	{Segment{6, 3}, TailTail{-2, +1}, Segment{5, 2}}, // case 65
	{Segment{6, 3}, TailTail{+2, +1}, Segment{1, 1}}, // case 66
	{Segment{6, 3}, TailTail{-2, -1}, Segment{5, 4}}, // case 67

	{Regions{Segment{3, 6}, Segment{13, 16}}, Head(+0), Segment{3, 3}},   // case 68
	{Regions{Segment{3, 6}, Segment{13, 16}}, Head(+7), Segment{17, 17}}, // case 69
	{Regions{Segment{3, 6}, Segment{13, 16}}, Tail(+0), Segment{16, 16}}, // case 70
	{Regions{Segment{3, 6}, Segment{13, 16}}, Tail(-7), Segment{2, 2}},   // case 71
	{Regions{Segment{13, 16}, Segment{3, 6}}, Head(+0), Segment{13, 13}}, // case 72
	{Regions{Segment{13, 16}, Segment{3, 6}}, Head(+7), Segment{7, 7}},   // case 73
	{Regions{Segment{13, 16}, Segment{3, 6}}, Tail(+0), Segment{6, 6}},   // case 74
	{Regions{Segment{13, 16}, Segment{3, 6}}, Tail(-7), Segment{12, 12}}, // case 75

	{Regions{Segment{3, 6}, Segment{13, 16}}, HeadTail{0, 0}, Regions{Segment{3, 6}, Segment{13, 16}}},  // case 76
	{Regions{Segment{3, 6}, Segment{13, 16}}, HeadTail{4, -4}, Segment{14, 14}},                         // case 77
	{Regions{Segment{3, 6}, Segment{13, 16}}, HeadHead{-2, 4}, Regions{Segment{1, 6}, Segment{13, 14}}}, // case 78
	{Regions{Segment{3, 6}, Segment{13, 16}}, TailTail{-4, 2}, Regions{Segment{5, 6}, Segment{13, 18}}}, // case 79
}

func TestRegionResize(t *testing.T) {
	for i, tt := range regionResizeTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := tt.in.Resize(tt.modifier)
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf(
					"resize by %s\n   in: %#v\n  out: %#v\n  exp: %#v",
					tt.modifier, tt.in, out, tt.out,
				)
			}
		})
	}
}

var regionLocateTests = []struct {
	in  Region
	out Sequence
}{
	{Segment{2, 6}, AsSequence("gcat")},
	{Segment{6, 2}, AsSequence("atgc")},
	{Regions{Segment{0, 2}, Segment{4, 6}}, AsSequence("atat")},
}

func TestRegionLocate(t *testing.T) {
	seq := AsSequence("atgcatgc")
	for i, tt := range regionLocateTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out, exp := tt.in.Locate(seq), tt.out
			testutils.Diff(t, string(out.Bytes()), string(exp.Bytes()))

			cmp := tt.in.Complement()
			if cmp.Len() != tt.in.Len() {
				t.Errorf("%s.Len() = %d, want %d", cmp, cmp.Len(), tt.in.Len())
			}

			if !reflect.DeepEqual(cmp.Complement(), tt.in) {
				t.Errorf(
					"%s.Complement() = %s, want %s",
					cmp, cmp.Complement(), tt.in,
				)
			}

			out = cmp.Locate(seq)
			exp = Apply(tt.out, Reverse, Complement)
			testutils.Diff(t, string(out.Bytes()), string(exp.Bytes()))
		})
	}
}

var bySegmentTests = [][]Segment{
	{{3, 13}, {4, 13}, {6, 14}, {6, 16}},
	{{13, 3}, {13, 4}, {14, 6}, {16, 6}},
}

func TestBySegment(t *testing.T) {
	for i, tt := range bySegmentTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			in := make([]Segment, len(tt))
			exp := make([]Segment, len(tt))
			out := make([]Segment, len(tt))

			copy(in, tt)
			copy(exp, tt)

			for reflect.DeepEqual(in, exp) {
				rand.Shuffle(len(in), func(i, j int) {
					in[i], in[j] = in[j], in[i]
				})
			}

			copy(out, in)
			sort.Sort(BySegment(out))
			if !reflect.DeepEqual(out, exp) {
				t.Errorf("sort.Sort(BySegment(%v)) = %v, want %v", in, out, exp)
			}
		})
	}
}

func TestMinimize(t *testing.T) {
	in := Regions{Segment{1, 3}, Segment{6, 9}, Segment{5, 3}, Segment{6, 8}, Segment{1, 3}}
	exp := []Segment{{1, 5}, {6, 9}}
	out := Minimize(in)
	if !reflect.DeepEqual(out, exp) {
		t.Errorf("Minimize(%#v) = %#v, want %#v", in, out, exp)
	}
}

var regionInvertLinearTests = []struct {
	in  Region
	n   int
	out []Region
}{
	{Segment{3, 5}, 7, []Region{Segment{0, 3}, Segment{5, 7}}},
}

func TestRegionInvertLinear(t *testing.T) {
	for i, tt := range regionInvertLinearTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := InvertLinear(tt.in, tt.n)
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf("InvertLinear(%#v) = %#v, want %#v", tt.in, out, tt.out)
			}
		})
	}
}

var regionInvertCircularTests = []struct {
	in  Region
	n   int
	out []Region
}{
	{Segment{3, 5}, 7, []Region{Regions{Segment{5, 7}, Segment{0, 3}}}},
	{Segment{0, 3}, 7, []Region{Segment{3, 7}}},
	{Segment{5, 7}, 7, []Region{Segment{0, 5}}},
}

func TestRegionInvertCircular(t *testing.T) {
	for i, tt := range regionInvertCircularTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := InvertCircular(tt.in, tt.n)
			if !reflect.DeepEqual(out, tt.out) {
				t.Errorf("InvertCircular(%#v) = %#v, want %#v", tt.in, out, tt.out)
			}
		})
	}
}

var testFeatureTable = Features{
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

var regionCropTests = []struct {
	r   Region
	out Features
}{
	// case 1
	{
		Segment{0, 150},
		Features{
			NewFeature("exon", Range(0, 42), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "alignment:Splign:2.1.0"},
			}),
			NewFeature("source", Partial3(0, 150), Props{
				[]string{"chromosome", "11"},
				[]string{"db_xref", "taxon:9606"},
				[]string{"map", "11p15.5"},
				[]string{"mol_type", "mRNA"},
				[]string{"organism", "Homo sapiens"},
			}),
			NewFeature("gene", Partial3(0, 150), Props{
				[]string{"db_xref", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "insulin"},
			}),
			NewFeature("exon", Partial3(42, 150), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "alignment:Splign:2.1.0"},
			}),
			NewFeature("sig_peptide", Range(59, 131), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "COORDINATES: ab initio prediction:SignalP:4.0"},
			}),
			NewFeature("CDS", Partial3(59, 150), Props{
				[]string{"codon_start", "1"},
				[]string{"db_xref", "CCDS:CCDS7729.1", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "proinsulin; preproinsulin"},
				[]string{"product", "insulin preproprotein"},
				[]string{"protein_id", "NP_000198.1"},
				[]string{"translation", "MALWMRLLPLLALLALWGPDPAAAFVNQHLCGSHLVEALYLVCG\nERGFFYTPKTRREAEDLQVGQVELGGGPGAGSLQPLALEGSLQKRGIVEQCCTSICSL\nYQLENYCN"},
			}),
			NewFeature("proprotein", Partial3(131, 150), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "proinsulin"},
			}),
			NewFeature("mat_peptide", Partial3(131, 150), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "insulin B chain"},
			}),
		},
	},

	// case 2
	{
		Segment{150, 0},
		Features{
			NewFeature("exon", Range(0, 42).Complement(), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "alignment:Splign:2.1.0"},
			}),
			NewFeature("source", Partial3(0, 150).Complement(), Props{
				[]string{"chromosome", "11"},
				[]string{"db_xref", "taxon:9606"},
				[]string{"map", "11p15.5"},
				[]string{"mol_type", "mRNA"},
				[]string{"organism", "Homo sapiens"},
			}),
			NewFeature("gene", Partial3(0, 150).Complement(), Props{
				[]string{"db_xref", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "insulin"},
			}),
			NewFeature("exon", Partial3(42, 150).Complement(), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "alignment:Splign:2.1.0"},
			}),
			NewFeature("sig_peptide", Range(59, 131).Complement(), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "COORDINATES: ab initio prediction:SignalP:4.0"},
			}),
			NewFeature("CDS", Partial3(59, 150).Complement(), Props{
				[]string{"codon_start", "1"},
				[]string{"db_xref", "CCDS:CCDS7729.1", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "proinsulin; preproinsulin"},
				[]string{"product", "insulin preproprotein"},
				[]string{"protein_id", "NP_000198.1"},
				[]string{"translation", "MALWMRLLPLLALLALWGPDPAAAFVNQHLCGSHLVEALYLVCG\nERGFFYTPKTRREAEDLQVGQVELGGGPGAGSLQPLALEGSLQKRGIVEQCCTSICSL\nYQLENYCN"},
			}),
			NewFeature("proprotein", Partial3(131, 150).Complement(), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "proinsulin"},
			}),
			NewFeature("mat_peptide", Partial3(131, 150).Complement(), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "insulin B chain"},
			}),
		},
	},

	// case 3
	{
		Regions{
			Segment{0, 150},
			Segment{300, 465},
		},
		Features{
			NewFeature("exon", Range(0, 42), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "alignment:Splign:2.1.0"},
			}),
			NewFeature("source", Partial3(0, 150), Props{
				[]string{"chromosome", "11"},
				[]string{"db_xref", "taxon:9606"},
				[]string{"map", "11p15.5"},
				[]string{"mol_type", "mRNA"},
				[]string{"organism", "Homo sapiens"},
			}),
			NewFeature("gene", Partial3(0, 150), Props{
				[]string{"db_xref", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "insulin"},
			}),
			NewFeature("exon", Partial3(42, 150), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "alignment:Splign:2.1.0"},
			}),
			NewFeature("sig_peptide", Range(59, 131), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "COORDINATES: ab initio prediction:SignalP:4.0"},
			}),
			NewFeature("CDS", Partial3(59, 150), Props{
				[]string{"codon_start", "1"},
				[]string{"db_xref", "CCDS:CCDS7729.1", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "proinsulin; preproinsulin"},
				[]string{"product", "insulin preproprotein"},
				[]string{"protein_id", "NP_000198.1"},
				[]string{"translation", "MALWMRLLPLLALLALWGPDPAAAFVNQHLCGSHLVEALYLVCG\nERGFFYTPKTRREAEDLQVGQVELGGGPGAGSLQPLALEGSLQKRGIVEQCCTSICSL\nYQLENYCN"},
			}),
			NewFeature("proprotein", Partial3(131, 150), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "proinsulin"},
			}),
			NewFeature("mat_peptide", Partial3(131, 150), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "insulin B chain"},
			}),
			NewFeature("mat_peptide", Partial5(150, 170), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "C-peptide"},
			}),
			NewFeature("proprotein", Partial5(150, 239), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "proinsulin"},
			}),
			NewFeature("CDS", Partial5(150, 242), Props{
				[]string{"codon_start", "1"},
				[]string{"db_xref", "CCDS:CCDS7729.1", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "proinsulin; preproinsulin"},
				[]string{"product", "insulin preproprotein"},
				[]string{"protein_id", "NP_000198.1"},
				[]string{"translation", "MALWMRLLPLLALLALWGPDPAAAFVNQHLCGSHLVEALYLVCG\nERGFFYTPKTRREAEDLQVGQVELGGGPGAGSLQPLALEGSLQKRGIVEQCCTSICSL\nYQLENYCN"},
			}),
			NewFeature("source", Partial5(150, 315), Props{
				[]string{"chromosome", "11"},
				[]string{"db_xref", "taxon:9606"},
				[]string{"map", "11p15.5"},
				[]string{"mol_type", "mRNA"},
				[]string{"organism", "Homo sapiens"},
			}),
			NewFeature("gene", Partial5(150, 315), Props{
				[]string{"db_xref", "GeneID:3630", "HGNC:HGNC:6081", "MIM:176730"},
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"note", "insulin"},
			}),
			NewFeature("exon", Partial5(150, 315), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"inference", "alignment:Splign:2.1.0"},
			}),
			NewFeature("mat_peptide", Range(176, 239), Props{
				[]string{"gene", "INS"},
				[]string{"gene_synonym", "IDDM; IDDM1; IDDM2; ILPR; IRDN; MODY10"},
				[]string{"product", "insulin A chain"},
			}),
		},
	},
}

func TestRegionCrop(t *testing.T) {
	for i, tt := range regionCropTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := tt.r.Crop(testFeatureTable)
			featuresDiff(t, out, tt.out)
		})
	}
}
