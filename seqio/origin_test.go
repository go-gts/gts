package seqio

import (
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func multiLineString(ss ...string) string {
	return strings.Join(ss, "\n") + "\n"
}

var originTests = []struct {
	in, out string
}{
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaa",
		"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa\n",
	},
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaat",
		multiLineString(
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat",
		),
	},
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaatcgaagtgga",
		multiLineString(
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtgga",
		),
	},
	{
		"gagttttatcgcttccatgacgcagaagttaacactttcggatatttctgatgagtcgaaaaattatcttgataaagcaggaattactactgcttgtttacgaattaaatcgaagtggac",
		multiLineString(
			"        1 gagttttatc gcttccatga cgcagaagtt aacactttcg gatatttctg atgagtcgaa",
			"       61 aaattatctt gataaagcag gaattactac tgcttgttta cgaattaaat cgaagtggac",
		),
	},
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
