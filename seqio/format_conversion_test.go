package seqio

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/testutils"
	"github.com/go-pars/pars"
)

func parseString(parser pars.Parser, s string) (gts.Sequence, error) {
	state := pars.FromString(s)
	result, err := parser.Parse(state)
	if err != nil {
		return GenBank{}, fmt.Errorf("parser returned %v\nBuffer:\n%q", err, string(result.Token))
	}
	return result.Value.(gts.Sequence), nil
}

func TestFormatConversion(t *testing.T) {
	s1 := testutils.ReadTestfile(t, "NC_001422.gb")
	s2 := testutils.ReadTestfile(t, "NC_001422.fasta")

	seq1, err := parseString(GenBankParser, s1)
	if err != nil {
		t.Error(err)
		return
	}

	seq2, err := parseString(FastaParser, s2)
	if err != nil {
		t.Error(err)
		return
	}

	testutils.Equals(t, bytes.ToUpper(seq1.Bytes()), bytes.ToUpper(seq2.Bytes()))
	formatter := NewFormatter(seq1, FastaFile)
	b := &strings.Builder{}
	n, err := formatter.WriteTo(b)
	if int(n) != len(s2) || err != nil {
		t.Errorf("formatter.WriteTo(builder) = (%d, %v), want (%d, nil)", n, err, len(s2))
	}
	out := b.String()
	testutils.Diff(t, strings.ToUpper(s2), strings.ToUpper(out))
}
