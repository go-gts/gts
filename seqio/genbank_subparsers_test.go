package seqio

import (
	"fmt"
	"testing"

	"github.com/go-pars/pars"
)

var genbankSubparsersTests = []struct {
	name   string
	parser pars.Parser
	pass   []string
	fail   []string
}{
	{
		"DBLink Parser",
		genbankDBLinkParser(&GenBank{}, 12),
		[]string{
			multiLineString(
				"DBLINK      BioProject: PRJNA14015",
				"            KEGG BRITE:  NC_001422",
			),
		},
		[]string{
			multiLineString(
				"DBLINK      BioProject: PRJNA14015",
				"            KEGG BRITE",
			),
		},
	},
	{
		"Contig Parser",
		genbankContigParser(&GenBank{}, 12),
		[]string{
			"CONTIG      join(U00096.3:1..4641652)",
		},
		[]string{
			"CONTIG      join(U00096.3)",
			"CONTIG      join(U00096.3:foo)",
			"CONTIG      join(U00096.3:1foo)",
			"CONTIG      join(U00096.3:1..foo)",
			"CONTIG      join(U00096.3:1..4641652",
		},
	},
	{
		"Origin Parser",
		makeGenbankOriginParser(120)(&GenBank{}, 12),
		[]string{
			multiLineString(
				"ORIGIN      ",
				"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
				"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac",
			),
		},
		[]string{
			multiLineString(
				"ORIGIN      ",
				"          gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
				"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac",
			),
			multiLineString(
				"ORIGIN      ",
				"        1g agttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
				"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac",
			),
			multiLineString(
				"ORIGIN      ",
				"        1  agttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
				"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac",
			),
		},
	},
}

func TestGenBankSubparsers(t *testing.T) {
	for _, tt := range genbankSubparsersTests {
		t.Run(fmt.Sprintf("%s pass tests", tt.name), func(t *testing.T) {
			for _, s := range tt.pass {
				state, result := pars.FromString(s), &pars.Result{}
				if err := tt.parser(state, result); err != nil {
					t.Errorf("%v while parsing:\n%s", err, s)
				}
			}
		})
		t.Run(fmt.Sprintf("%s fail tests", tt.name), func(t *testing.T) {
			for _, s := range tt.fail {
				state, result := pars.FromString(s), &pars.Result{}
				if tt.parser(state, result) == nil {
					t.Errorf("expected error while parsing:\n%s", s)
				}
			}
		})
	}
}
