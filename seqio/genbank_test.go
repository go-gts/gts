package seqio

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-pars/pars"
)

func TestGenBankHeaderSlice(t *testing.T) {
	b := &bytes.Buffer{}

	in := testutils.ReadTestfile(t, "NC_001422.gb")
	state := pars.FromString(in)
	stream := NewGenBankIOStream(state, b)

	exp := testutils.ReadTestfile(t, "NC_001422_part.gb")

	err := stream.ForEach(func(i int, header interface{}, ff gts.Features) (Callback, error) {
		switch v := header.(type) {
		case GenBankHeader:
			if err := stream.PushHeader(v.Slice(2379, 2512)); err != nil {
				return nil, err
			}
			if err := stream.PushFeatures(gts.Segment{2379, 2512}.Crop(ff)); err != nil {
				return nil, err
			}

		default:
			t.Errorf("header.(type) = %T, want %T", header, GenBankHeader{})
			return nil, nil
		}

		return func(seq gts.Sequence) error {
			return stream.PushSequence(gts.Slice(seq, 2379, 2512))
		}, nil
	})

	if err != nil {
		t.Errorf("stream.ForEach: %v", err)
	}

	testutils.DiffLine(t, b.String(), exp)
}

var originTests = []struct {
	in  string
	out string
}{
	// case 1
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaa",
		"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa\n",
	},

	// case 2
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaat",
		multiLineString(
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat",
			"",
		),
	},

	// case 3
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaatcgaagtgga",
		multiLineString(
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtgga",
			"",
		),
	},

	// case 4
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaatcgaagtggac",
		multiLineString(
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac",
			"",
		),
	},

	// case 5
	{"", ""},
}

func TestOrigin(t *testing.T) {
	for _, tt := range originTests {
		o := NewOrigin([]byte(tt.in))
		out := o.String()
		if out != tt.out {
			testutils.DiffLine(t, out, tt.out)
		}
		if o.Len() != len(tt.in) {
			t.Errorf("o.Len() = %d, want %d", o.Len(), len(tt.in))
		}

		out = string(o.Bytes())
		if out != tt.in {
			testutils.Diff(t, out, tt.in)
		}
		if o.Len() != len(tt.in) {
			t.Errorf("o.Len() = %d, want %d", o.Len(), len(tt.in))
		}

		out = o.String()
		if out != tt.out {
			testutils.DiffLine(t, out, tt.out)
		}
		out = string(o.Buffer)
		if out != tt.in {
			testutils.Diff(t, out, tt.in)
		}
	}
}

var genbankConfigIOFailTests = []string{
	"FOO",                                  // case 1
	"CONTIG      join",                     // case 2
	"CONTIG      join(U00096.3",            // case 3
	"CONTIG      join(U00096.3:",           // case 4
	"CONTIG      join(U00096.3:1",          // case 5
	"CONTIG      join(U00096.3:1..",        // case 6
	"CONTIG      join(U00096.3:1..4641652", // case 7
}

func TestContigIO(t *testing.T) {
	var seq gts.Sequence
	parser := genbankContigParser(&seq, defaultGenBankIndentLength)

	t.Run("Pass", func(t *testing.T) {
		in := "CONTIG      join(U00096.3:1..4641652)\n"
		state := pars.FromString(in)
		result := &pars.Result{}
		if err := parser(state, result); err != nil {
			t.Errorf("while parsing:\n%q\ngot: %v", in, err)
		}
		switch contig := seq.(type) {
		case Contig:
			testutils.Equals(t, contig.Len(), 4641652)
			out := fmtGenBankField("CONTIG", contig.String())
			testutils.Diff(t, out, in)
		default:
			t.Errorf("expected result to be type Contig: got %T", contig)
		}
	})

	testutils.Equals(t, Contig{}.String(), "")

	for i, tt := range genbankConfigIOFailTests {
		t.Run(fmt.Sprintf("fail case %d", i+1), func(t *testing.T) {
			state := pars.FromString(tt)
			result := &pars.Result{}
			if err := parser(state, result); err == nil {
				t.Errorf("while parsing:%q\nexpected error", tt)
			}
		})
	}

	testutils.Panics(t, func() {
		Contig{}.Bytes()
	})
}

type genbankHeaderFieldParserCase struct {
	in  string
	out *GenBankHeader
}

var genbankHeaderFieldParserGeneratorTests = []struct {
	gen   func(header *GenBankHeader, depth int) pars.Parser
	cases []genbankHeaderFieldParserCase
}{
	// case 1
	{
		genbankExtraFieldParser,
		[]genbankHeaderFieldParserCase{
			{fmtGenBankField("FOO", "BAR\nBAZ"), &GenBankHeader{Extra: []ExtraField{{"FOO", "BAR\nBAZ"}}}},
			{"", nil},
			{"foo", nil},
			{"FOO BAR", nil},
		},
	},

	// case 2
	{
		genbankDefinitionParser,
		[]genbankHeaderFieldParserCase{
			{
				fmtGenBankField("DEFINITION", "Coliphage phi-X174, complete genome."),
				&GenBankHeader{Definition: "Coliphage phi-X174, complete genome"},
			},
			{
				fmtGenBankField("DEFINITION", "Coliphage phi-X174, complete genome"),
				nil,
			},
			{
				"FOO",
				nil,
			},
		},
	},

	// case 3
	{
		genbankAccessionParser,
		[]genbankHeaderFieldParserCase{
			{
				fmtGenBankField("ACCESSION", "NC_001422"),
				&GenBankHeader{Accession: "NC_001422"},
			},
			{
				fmtGenBankField("ACCESSION", "NC_001422 REGION: 2380..2512"),
				&GenBankHeader{Accession: "NC_001422", Region: gts.Segment{2379, 2512}},
			},
			{
				fmtGenBankField("ACCESSION", "NC_001422 REGION: FOO"),
				&GenBankHeader{Accession: "NC_001422 REGION: FOO"},
			},
			{
				fmtGenBankField("ACCESSION", "NC_001422 REGION: join(3981..5386,1..136)"),
				&GenBankHeader{Accession: "NC_001422 REGION: join(3981..5386,1..136)"},
			},
			{
				"FOO",
				nil,
			},
		},
	},

	// case 4
	{
		genbankVersionParser,
		[]genbankHeaderFieldParserCase{
			{
				fmtGenBankField("VERSION", "NC_001422.1"),
				&GenBankHeader{Version: "NC_001422.1"},
			},
			{
				"FOO",
				nil,
			},
		},
	},

	// case 5
	{
		genbankDBLinkParser,
		[]genbankHeaderFieldParserCase{
			{
				fmtGenBankField("DBLINK", "BioProject: PRJNA14015\nKEGG BRITE: NC_001422"),
				&GenBankHeader{DBLink: gts.Props{
					[]string{"BioProject", "PRJNA14015"},
					[]string{"KEGG BRITE", "NC_001422"},
				}},
			},
			{"", nil},
			{fmtGenBankField("DBLINK", "BioProject\nKEGG BRITE: NC_001422"), nil},
			{fmtGenBankField("DBLINK", "BioProject: PRJNA14015\nKEGG BRITE"), nil},
		},
	},

	// case 6
	{
		genbankKeywordsParser,
		[]genbankHeaderFieldParserCase{
			{
				fmtGenBankField("KEYWORDS", "RefSeq."),
				&GenBankHeader{Keywords: []string{"RefSeq"}},
			},
			{"", nil},
		},
	},

	// case 7
	{
		genbankSourceParser,
		[]genbankHeaderFieldParserCase{
			{
				strings.Join([]string{
					fmtGenBankField("SOURCE", "Escherichia virus phiX174"),
					fmtGenBankField("  ORGANISM", multiLineString(
						"Escherichia virus phiX174",
						"Viruses; Monodnaviria; Sangervirae; Phixviricota;",
						"Malgrandaviricetes; Petitvirales; Microviridae; Bullavirinae;",
						"Sinsheimervirus.",
					)),
				}, ""),
				&GenBankHeader{Source: Organism{
					"Escherichia virus phiX174",
					"Escherichia virus phiX174",
					[]string{
						"Viruses",
						"Monodnaviria",
						"Sangervirae",
						"Phixviricota",
						"Malgrandaviricetes",
						"Petitvirales",
						"Microviridae",
						"Bullavirinae",
						"Sinsheimervirus",
					},
				}},
			},
			{"", nil},
			{
				strings.Join([]string{
					fmtGenBankField("FOO", "Escherichia virus phiX174"),
					fmtGenBankField("  ORGANISM", multiLineString(
						"Escherichia virus phiX174",
						"Viruses; Monodnaviria; Sangervirae; Phixviricota;",
						"Malgrandaviricetes; Petitvirales; Microviridae; Bullavirinae;",
						"Sinsheimervirus.",
					)),
				}, ""),
				nil,
			},
			{
				strings.Join([]string{
					fmtGenBankField("SOURCE", "Escherichia virus phiX174"),
					fmtGenBankField("ORGANISM", "foo"),
				}, ""),
				nil,
			},
			{
				strings.Join([]string{
					fmtGenBankField("SOURCE", "Escherichia virus phiX174"),
					fmtGenBankField("  FOO", "foo"),
				}, ""),
				nil,
			},
			{
				strings.Join([]string{
					fmtGenBankField("SOURCE", "Escherichia virus phiX174"),
					"  ORGANISM foo",
				}, ""),
				nil,
			},
		},
	},

	// case 8
	{
		genbankReferenceParser,
		[]genbankHeaderFieldParserCase{
			{
				strings.Join([]string{
					fmtGenBankField("REFERENCE", "1  (bases 2380 to 2512; 2593 to 2786; 2788 to 2947)"),
					fmtGenBankField("  AUTHORS", "Air,G.M., Els,M.C., Brown,L.E., Laver,W.G. and Webster,R.G."),
					fmtGenBankField("  TITLE", multiLineString(
						"Location of antigenic sites on the three-dimensional structure of",
						"the influenza N2 virus neuraminidase",
					)),
					fmtGenBankField("  JOURNAL", "Virology 145 (2), 237-248 (1985)"),
					fmtGenBankField("   PUBMED", "2411049"),
					fmtGenBankField("  REMARK", "Reference comment."),
				}, ""),
				&GenBankHeader{References: []Reference{{
					Number:  1,
					Info:    "(bases 2380 to 2512; 2593 to 2786; 2788 to 2947)",
					Authors: "Air,G.M., Els,M.C., Brown,L.E., Laver,W.G. and Webster,R.G.",
					Title: multiLineString(
						"Location of antigenic sites on the three-dimensional structure of",
						"the influenza N2 virus neuraminidase",
					),
					Journal: "Virology 145 (2), 237-248 (1985)",
					Xref: map[string]string{
						"PUBMED": "2411049",
					},
					Comment: "Reference comment.",
				}}},
			},
			{
				strings.Join([]string{
					fmtGenBankField("REFERENCE", "24 (bases 1 to 5386)"),
					fmtGenBankField("  CONSRTM", "NCBI Genome Project"),
					fmtGenBankField("  TITLE", "Direct Submission"),
					fmtGenBankField("  JOURNAL", multiLineString(
						"Submitted (06-JUL-2018) National Center for Biotechnology",
						"Information, NIH, Bethesda, MD 20894, USA",
					)),
				}, ""),
				&GenBankHeader{References: []Reference{{
					Number: 24,
					Info:   "(bases 1 to 5386)",
					Group:  "NCBI Genome Project",
					Title:  "Direct Submission",
					Journal: multiLineString(
						"Submitted (06-JUL-2018) National Center for Biotechnology",
						"Information, NIH, Bethesda, MD 20894, USA",
					),
				}}},
			},
			{"", nil},
			{fmtGenBankField("REFERENCE", "foo"), nil},
		},
	},

	// case 9
	{
		genbankCommentParser,
		[]genbankHeaderFieldParserCase{
			{
				fmtGenBankField("COMMENT", multiLineString(
					"PROVISIONAL REFSEQ: This record has not yet been subject to final",
					"NCBI review. The reference sequence is identical to J02482.",
					"[8]  intermittent sequences.",
					"[15]  review; discussion of complete genome.",
					"Double checked with sumex tape.",
					"Single-stranded circular DNA which codes for eleven proteins.",
					"Replicative form is duplex, icosahedron, related to s13 & g4. [21]",
					"indicates that mitomycin C reduced with sodium borohydride induced",
					"heat-labile sites in DNA most preferentially at dinucleotide",
					"sequence 'gt' (especially 'Pu-g-t').",
					"Bacteriophage phi-X174 single stranded DNA molecules were",
					"irradiated with near UV light in the presence of promazine",
					"derivatives, after priming with restriction fragments or synthetic",
					"primers [22].  The resulting DNA fragments were used as templates",
					"for in vitro complementary chain synthesis by E.coli DNA polymerase",
					"I [22].  More than 90% of the observed chain terminations were",
					"mapped one nucleotide before a guanine residue [22].  Photoreaction",
					"occurred more predominantly with guanine residues localized in",
					"single-stranded parts of the genome [22].  These same guanine",
					"residues could also be damaged when the reaction was performed in",
					"the dark, in the presence of promazine cation radicals [22].",
					"COMPLETENESS: full length.",
				)),
				&GenBankHeader{
					Comments: []string{multiLineString(
						"PROVISIONAL REFSEQ: This record has not yet been subject to final",
						"NCBI review. The reference sequence is identical to J02482.",
						"[8]  intermittent sequences.",
						"[15]  review; discussion of complete genome.",
						"Double checked with sumex tape.",
						"Single-stranded circular DNA which codes for eleven proteins.",
						"Replicative form is duplex, icosahedron, related to s13 & g4. [21]",
						"indicates that mitomycin C reduced with sodium borohydride induced",
						"heat-labile sites in DNA most preferentially at dinucleotide",
						"sequence 'gt' (especially 'Pu-g-t').",
						"Bacteriophage phi-X174 single stranded DNA molecules were",
						"irradiated with near UV light in the presence of promazine",
						"derivatives, after priming with restriction fragments or synthetic",
						"primers [22].  The resulting DNA fragments were used as templates",
						"for in vitro complementary chain synthesis by E.coli DNA polymerase",
						"I [22].  More than 90% of the observed chain terminations were",
						"mapped one nucleotide before a guanine residue [22].  Photoreaction",
						"occurred more predominantly with guanine residues localized in",
						"single-stranded parts of the genome [22].  These same guanine",
						"residues could also be damaged when the reaction was performed in",
						"the dark, in the presence of promazine cation radicals [22].",
						"COMPLETENESS: full length.",
					)},
				},
			},
			{"", nil},
		},
	},
}

func TestGenBankHeaderFieldParserGenerators(t *testing.T) {
	for _, tt := range genbankHeaderFieldParserGeneratorTests {
		ptr := reflect.ValueOf(tt.gen).Pointer()
		fn := runtime.FuncForPC(ptr)
		name := path.Ext(fn.Name())[1:]

		t.Run(name, func(t *testing.T) {

			for i, c := range tt.cases {
				out := &GenBankHeader{}
				parser := tt.gen(out, defaultGenBankIndentLength)

				t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
					state := pars.FromString(c.in)
					result := &pars.Result{}

					err := parser(state, result)

					switch c.out {
					case nil:
						if err == nil {
							t.Errorf("while parsing:\n%q\nexpected error", c.in)
						}

					default:
						if err != nil {
							t.Errorf("while parsing:\n%q\ngot: %v", c.in, err)
						}
						testutils.Equals(t, out, c.out)
					}
				})
			}
		})
	}
}

var genbankFeatureParserTests = []struct {
	in  string
	out gts.Features
}{
	// case 1
	{
		multiLineString(
			"FEATURES             Location/Qualifiers",
			"     source          1..465",
			"                     /organism=\"Homo sapiens\"",
			"                     /mol_type=\"mRNA\"",
			"                     /db_xref=\"taxon:9606\"",
			"                     /chromosome=\"11\"",
			"                     /map=\"11p15.5\"",
		),
		gts.Features{
			{
				Key: "source",
				Loc: gts.Range(0, 465),
				Props: gts.Props{
					[]string{"organism", "Homo sapiens"},
					[]string{"mol_type", "mRNA"},
					[]string{"db_xref", "taxon:9606"},
					[]string{"chromosome", "11"},
					[]string{"map", "11p15.5"},
				},
			},
		},
	},

	// case 2
	{
		multiLineString(
			"FEATURES             Location/Qualifiers",
			"     source          1..5386",
			"                     /organism=\"Escherichia virus phiX174\"",
			"                     /mol_type=\"genomic DNA\"",
			"                     /db_xref=\"taxon:10847\"",
		),
		gts.Features{
			{
				Key: "source",
				Loc: gts.Range(0, 5386),
				Props: gts.Props{
					[]string{"organism", "Escherichia virus phiX174"},
					[]string{"mol_type", "genomic DNA"},
					[]string{"db_xref", "taxon:10847"},
				},
			},
		},
	},

	// case 3
	{"", nil},

	// case 4
	{multiLineString("FEATURES             Location/Qualifiers", "     source          "), nil},
}

func TestGenBankFeatureParser(t *testing.T) {
	ff := gts.Features{}
	parser := genbankFeatureParser(&ff)
	for i, tt := range genbankFeatureParserTests {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			state := pars.FromString(tt.in)
			result := &pars.Result{}
			err := parser(state, result)
			switch tt.out {
			case nil:
				if err == nil {
					t.Errorf("genbankFeatureParser expected error")
				}

			default:
				if err != nil {
					t.Errorf("genbankFeatureParser: %v", err)
				}
				switch ff := result.Value.(type) {
				case gts.Features:
					testutils.Equals(t, ff, tt.out)
				default:
					t.Errorf("expected result to be type gts.Features: got %T", ff)
				}
			}
		})
	}
}

var genbankOriginParserPassTests = []struct {
	in  string
	out string
}{
	// case 1
	{
		multiLineString(
			"ORIGIN      ",
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac",
			"",
		),
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaatcgaagtggac",
	},

	// case 2
	{

		multiLineString(
			"ORIGIN      ",
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa\r",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac\r",
			"\r",
		),
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaatcgaagtggac",
	},
}

var genbankOriginParserFailTests = []string{
	// case 1
	"ORIG",

	// case 2
	"ORIGIN      ",

	// case 3
	multiLineString(
		"ORIGIN      ",
		"        ? gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa\r",
		"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac\r",
		"\r",
	),

	// case 4
	multiLineString(
		"ORIGIN      ",
		"        1?gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa\r",
		"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac\r",
		"\r",
	),

	// case 5
	multiLineString(
		"ORIGIN      ",
		"        1  agttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa\r",
		"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac\r",
		"\r",
	),
}

func TestGenBankOriginParser(t *testing.T) {
	var seq gts.Sequence
	parser := makeGenbankOriginParser(&seq, 120, defaultGenBankIndentLength)

	t.Run("Pass", func(t *testing.T) {
		for i, tt := range genbankOriginParserPassTests {
			t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
				state := pars.FromString(tt.in)
				result := &pars.Result{}
				if err := parser(state, result); err != nil {
					t.Errorf("while parsing\n%q\ngot: %v", tt, err)
				}
				switch origin := seq.(type) {
				case *Origin:
					testutils.Equals(t, origin.Len(), 120)
					testutils.Diff(t, string(origin.Bytes()), tt.out)
				default:
					t.Errorf("expected result to be type *Origin: got %T", origin)
				}
			})
		}
	})

	t.Run("Fail", func(t *testing.T) {
		for i, tt := range genbankOriginParserFailTests {
			t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
				state := pars.FromString(tt)
				result := &pars.Result{}
				if err := parser(state, result); err == nil {
					t.Errorf("while parsing\n%q\nexpected error", tt)
				}
			})
		}
	})
}

var fmtGenBankFieldTests = []struct {
	name  string
	value string
	out   string
}{
	{"FOO", "foo", "FOO         foo\n"},                       // case 1
	{"FOO", "foo\nbar", "FOO         foo\n            bar\n"}, // case 2
}

func TestFmtGenBankField(t *testing.T) {
	for _, tt := range fmtGenBankFieldTests {
		out := fmtGenBankField(tt.name, tt.value)
		testutils.DiffLine(t, out, tt.out)
	}
}
func TestGenBankParser(t *testing.T) {
	files := []string{
		"NC_001422.gb",
		"NC_001422_part.gb",
		"NC_000913.3.min.gb",
		"pBAT5.txt",
	}

	for i, file := range files {
		testutils.RunCase(t, i, func(t *testing.T) {
			in := testutils.ReadTestfile(t, file)
			state := pars.FromString(in)
			parser := pars.AsParser(GenBankParser)

			result, err := parser.Parse(state)
			if err != nil {
				t.Errorf("in file %q, parser returned %v\nBuffer:\n%s", file, err, string(result.Token))
				return
			}

			if rec, ok := result.Value.(Record); !ok {
				t.Errorf("result.Value.(type) = %T, want %T", rec, Record{})
			}
		})
	}
}

var genbankParserFailTests = []string{
	// case 1
	"",

	// case 2
	"NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",

	// case 3
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     topology PHG 06-JUL-2018",
		"foo",
	),

	// case 4
	multiLineString(
		"LOCUS       NC_001422               5386 bp    foo     topology PHG 06-JUL-2018",
		"foo",
	),

	// case 5
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"foo",
	),

	// case 6
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"DEFINITION",
	),

	// case 7
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"DEFINITION ",
	),

	// case 8
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"DEFINITION  Coliphage phi-X174, complete genome",
	),

	// case 9
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"DBLINK      FOO",
	),

	// case 10
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"SOURCE      Escherichia virus phiX174",
		"  ORGANISM Escherichia virus phiX174",
	),

	// case 11
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"REFERENCE   ",
	),

	// case 12
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"REFERENCE   1",
	),

	// case 13
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"REFERENCE   1  (bases 2380 to 2512; 2593 to 2786; 2788 to 2947)",
		"  AUTHORS  Air,G.M., Els,M.C., Brown,L.E., Laver,W.G. and Webster,R.G.",
	),

	// case 14
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"FEATURES             Location/Qualifiers",
	),

	// case 15
	multiLineString(
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020",
		"ORIGIN      \n",
	),

	// case 16
	multiLineString(
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020",
		"ORIGIN      ",
		"       1 gagttttatc gcttccatga",
	),

	// case 17
	multiLineString(
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020",
		"ORIGIN      ",
		"        1 gagttttatcgcttccatga",
	),

	// case 18
	multiLineString(
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020",
		"ORIGIN      ",
		"        1  gagttttatc gcttccatga",
	),

	// case 19
	multiLineString(
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020",
		"CONTIG      ",
	),

	// case 20
	multiLineString(
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
		"FOO         ",
	),
}

func TestGenBankParserFail(t *testing.T) {
	parser := pars.AsParser(GenBankParser)
	for _, in := range genbankParserFailTests {
		state := pars.FromString(in)
		if err := parser(state, pars.Void); err == nil {
			t.Errorf("while parsing`\n%q\n`: expected error", in)
			return
		}
	}
}

func generateGenBankTestRecords() []Record {
	length := 60
	records := make([]Record, 3)

	seqs := []gts.Sequence{
		NewOrigin([]byte(StringWithCharset("atgc", length))),
		Contig{"TESTCONTIG", gts.Segment{0, length}},
		gts.AsSequence(StringWithCharset("atgc", length)),
	}

	for i, seq := range seqs {
		header := GenBankHeader{
			LocusName: fmt.Sprintf("TEST%06d", i),
			Molecule:  gts.DNA,
			Topology:  gts.Linear,
			Division:  "CON",
			Date:      Date{Year: 1992, Month: time.April, Day: 8},

			Definition: "Test sequence",
			Accession:  fmt.Sprintf("TEST%06d", i),
			Version:    fmt.Sprintf("TEST%06d", i),
			DBLink: gts.Props{
				[]string{"BioProject", "PRJNA14015"},
				[]string{"KEGG BRITE", "NC_001422"},
			},
			Keywords: []string{"Test", "Foo"},
			Source: Organism{
				Species: "Foo",
				Name:    "Foo",
				Taxon:   []string{"Foo", "Bar", "Baz"},
			},
			References: []Reference{
				{
					Number:  1,
					Authors: "Foo",
					Title:   "Foo",
					Journal: "Foo",
					Xref:    map[string]string{"PUBMED": "FOO"},
					Comment: "Foo",
				},
				{
					Number:  2,
					Info:    "(sites)",
					Group:   "Foo",
					Title:   "Foo",
					Journal: "Foo",
					Xref:    map[string]string{"PUBMED": "FOO"},
					Comment: "Foo",
				},
			},
			Comments: []string{"Foo"},
			Extra: []ExtraField{
				{"WARNING", "Test sequence data."},
			},

			Region: gts.Segment{0, length},
		}

		ff := []gts.Feature{
			{
				Key:   "source",
				Loc:   gts.Range(0, length),
				Props: gts.Props{},
			},
		}

		records[i] = Record{header, ff, seq}
	}

	return records
}

func TestGenBankIOStream(t *testing.T) {
	buf := &bytes.Buffer{}
	state := pars.NewState(buf)

	stream := NewGenBankIOStream(state, buf)
	genbankTestRecords := generateGenBankTestRecords()

	for _, rec := range genbankTestRecords {
		if err := stream.PushHeader(rec.Header); err != nil {
			t.Errorf("stream.PushHeader: %v", err)
		}
		if err := stream.PushFeatures(rec.Features); err != nil {
			t.Errorf("stream.PushFeatures: %v", err)
		}
		if err := stream.PushSequence(rec.Sequence); err != nil {
			t.Errorf("stream.PushSequence: %v", err)
		}
	}

	if stream.ForEach(func(i int, header interface{}, ff gts.Features) (Callback, error) {
		return nil, errors.New("error")
	}) == nil {
		t.Error("stream.ForEach: expected error")
	}

	if stream.ForEach(func(i int, header interface{}, ff gts.Features) (Callback, error) {
		return func(seq gts.Sequence) error {
			return errors.New("error")
		}, nil
	}) == nil {
		t.Error("stream.ForEach: expected error")
	}

	manip := func(i int, header interface{}, ff gts.Features) (Callback, error) {
		testutils.Equals(t, header, genbankTestRecords[i].Header)
		testutils.Equals(t, ff, genbankTestRecords[i].Features)
		return func(seq gts.Sequence) error {
			if _, ok := genbankTestRecords[i].Sequence.(gts.Seq); ok {
				seq = gts.Seq(seq.Bytes())
			}
			testutils.Equals(t, seq, genbankTestRecords[i].Sequence)
			return nil
		}, nil
	}

	if err := stream.ForEach(manip); err != nil {
		t.Errorf("stream.ForEach: %v", err)
	}

	if stream.PushHeader("foo") == nil {
		t.Errorf("stream.PushHeader(%q): expected error", "foo")
	}

	if stream.PushSequence(Contig{}) == nil {
		t.Errorf("stream.PushSequence(%q): expected error", Contig{})
	}
}
