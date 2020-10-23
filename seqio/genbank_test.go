package seqio

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/go-ascii/ascii"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-pars/pars"
)

func formatGenBankHelper(t *testing.T, seq gts.Sequence, in string) {
	t.Helper()
	b := strings.Builder{}
	f := GenBankFormatter{seq}
	n, err := f.WriteTo(&b)
	if int(n) != len([]byte(in)) || err != nil {
		t.Errorf("f.WriteTo(&b) = (%d, %v), want %d, nil", n, err, len(in))
	}
	testutils.DiffLine(t, in, b.String())
}

func TestGenBankFields(t *testing.T) {
	var (
		locusName = "LocusName"
		accession = "Accession"
		version   = "Version"
	)

	info := GenBankFields{}

	if info.ID() != "" {
		t.Errorf("info.ID() = %q, want %q", info.ID(), "")
	}
	info.LocusName = locusName
	if info.ID() != locusName {
		t.Errorf("info.ID() = %q, want %q", info.ID(), locusName)
	}
	info.Accession = accession
	if info.ID() != accession {
		t.Errorf("info.ID() = %q, want %q", info.ID(), accession)
	}
	info.Version = version
	if info.ID() != version {
		t.Errorf("info.ID() = %q, want %q", info.ID(), version)
	}
}

func TestGenBankWithInterface(t *testing.T) {
	length := 100

	info := GenBankFields{
		LocusName: "LOCUS_NAME",
		Molecule:  gts.DNA,
		Topology:  gts.Linear,
		Division:  "UNA",
		Date:      FromTime(time.Now()),

		Definition: "Sample sequence",
		Accession:  "ACCESSION",
		Version:    "VERSION",
		Source: Organism{
			Species: "Genus species",
			Name:    "Name",
			Taxon:   []string{"Kingdom", "Phylum", "Class", "Order", "Family", "Genus", "species"},
		},
	}

	p := []byte(strings.Repeat("atgc", length))
	qfs := gts.Values{}
	qfs.Add("organism", "Genus species")
	qfs.Add("mol_type", "Genomic DNA")
	loc := gts.Range(0, len(p))
	ff := []gts.Feature{
		{
			Key:        "source",
			Location:   loc,
			Qualifiers: qfs,
		},
	}

	in := GenBank{GenBankFields{}, nil, NewOrigin(nil)}
	out := gts.WithInfo(in, info)
	testutils.Equals(t, out, GenBank{info, nil, NewOrigin(nil)})

	out = gts.WithFeatures(in, ff)
	testutils.Equals(t, out, GenBank{GenBankFields{}, ff, NewOrigin(nil)})

	out = gts.WithBytes(in, p)
	testutils.Equals(t, out, GenBank{GenBankFields{}, nil, NewOrigin(p)})

	out = gts.WithInfo(in, "info")
	testutils.Equals(t, out, gts.New("info", nil, nil))

	out = gts.WithTopology(in, gts.Circular)
	top := out.(GenBank).Fields.Topology
	if top != gts.Circular {
		t.Errorf("topology is %q, expected %q", top, gts.Circular)
	}
}

func TestGenBankSlice(t *testing.T) {
	in := testutils.ReadTestfile(t, "NC_001422.gb")
	state := pars.FromString(in)
	parser := pars.AsParser(GenBankParser)

	exp := testutils.ReadTestfile(t, "NC_001422_part.gb")

	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("parser returned %v\nBuffer:\n%q", err, string(result.Token))
	}

	switch seq := result.Value.(type) {
	case GenBank:
		seq = gts.Slice(seq, 2379, 2512).(GenBank)
		formatGenBankHelper(t, seq, exp)

	default:
		t.Errorf("result.Value.(type) = %T, want %T", seq, GenBank{})
	}
}

func TestGenBankIO(t *testing.T) {
	files := []string{
		"NC_001422.gb",
		"NC_000913.3.min.gb",
	}
	for _, file := range files {
		in := testutils.ReadTestfile(t, file)
		state := pars.FromString(in)
		parser := pars.AsParser(GenBankParser)

		result, err := parser.Parse(state)
		if err != nil {
			t.Errorf("in file %q, parser returned %v\nBuffer:\n%q", file, err, string(result.Token))
			return
		}

		switch seq := result.Value.(type) {
		case GenBank:
			formatGenBankHelper(t, &seq, in)
			cpy := gts.New(seq.Info(), seq.Features(), seq.Bytes())
			formatGenBankHelper(t, &cpy, in)

		default:
			t.Errorf("result.Value.(type) = %T, want %T", seq, GenBank{})
		}
	}
}

func TestGenBankParser(t *testing.T) {
	files := []string{
		"NC_001422.gb",
		"pBAT5.txt",
	}

	for _, file := range files {
		in := testutils.ReadTestfile(t, file)
		state := pars.FromString(in)
		parser := pars.AsParser(GenBankParser)

		result, err := parser.Parse(state)
		if err != nil {
			t.Errorf("in file %q, parser returned %v\nBuffer:\n%q", file, err, string(result.Token))
			return
		}

		switch seq := result.Value.(type) {
		case GenBank:
			data := seq.Bytes()
			if len(data) != gts.Len(seq) {
				t.Errorf("in file %s: len(data) = %d, want %d", file, len(data), gts.Len(seq))
				return
			}
			if seq.Info() == nil {
				t.Errorf("in file %s: seq.Info() is nil", file)
				return
			}
			if seq.Features() == nil {
				t.Errorf("in file %s: seq.Features() is nil", file)
				return
			}
			for i, c := range data {
				if !ascii.IsLetterFilter(c) {
					t.Errorf("in file %s: origin contains `%U` at byte %d, expected a sequence character", file, c, i+1)
					return
				}
			}

		default:
			t.Errorf("result.Value.(type) = %T, want %T", seq, GenBank{})
		}
	}
}

var genbankIOFailTests = []string{
	"",
	"NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     topology PHG 06-JUL-2018\n" +
		"foo",
	"" +
		"LOCUS       NC_001422               5386 bp    foo     topology PHG 06-JUL-2018\n" +
		"foo",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"foo",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"DEFINITION",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"DEFINITION ",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"DEFINITION  Coliphage phi-X174, complete genome",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"DBLINK      FOO",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"SOURCE      Escherichia virus phiX174\n" +
		"  ORGANISM Escherichia virus phiX174",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"REFERENCE   ",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"REFERENCE   1",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"REFERENCE   1  (bases 2380 to 2512; 2593 to 2786; 2788 to 2947)\n" +
		"  AUTHORS  Air,G.M., Els,M.C., Brown,L.E., Laver,W.G. and Webster,R.G.",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"FEATURES             Location/Qualifiers",
	"" +
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020\n" +
		"ORIGIN      \n",
	"" +
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020\n" +
		"ORIGIN      \n" +
		"       1 gagttttatc gcttccatga",
	"" +
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020\n" +
		"ORIGIN      \n" +
		"        1 gagttttatcgcttccatga",
	"" +
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020\n" +
		"ORIGIN      \n" +
		"        1  gagttttatc gcttccatga",
	"" +
		"LOCUS       TEST_DATA                 20 bp    DNA     linear   UNA 14-MAY-2020\n" +
		"CONTIG      ",
	"" +
		"LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018\n" +
		"FOO         ",
}

func TestGenBankIOFail(t *testing.T) {
	parser := pars.AsParser(GenBankParser)
	for _, in := range genbankIOFailTests {
		state := pars.FromString(in)
		if err := parser(state, pars.Void); err == nil {
			t.Errorf("while parsing`\n%s\n`: expected error", in)
			return
		}
	}

	w := bytes.Buffer{}
	n, err := GenBankFormatter{gts.New(nil, nil, nil)}.WriteTo(&w)
	if n != 0 || err == nil {
		t.Errorf("formatting an empty Sequence should return an error")
		return
	}
}
