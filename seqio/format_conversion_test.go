package seqio

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
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
	testutils.DiffLine(t, strings.ToUpper(s2), strings.ToUpper(out))
}

func TestSliceToFasta(t *testing.T) {
	in := testutils.ReadTestfile(t, "NC_001422.gb")
	state := pars.FromString(in)
	parser := pars.AsParser(GenBankParser)

	exp := testutils.ReadTestfile(t, "NC_001422_part.fasta")

	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("parser returned %v\nBuffer:\n%q", err, string(result.Token))
	}

	switch seq := result.Value.(type) {
	case GenBank:
		seq = gts.Slice(seq, 2379, 2512).(GenBank)
		formatter := NewFormatter(seq, FastaFile)
		b := &strings.Builder{}
		n, err := formatter.WriteTo(b)
		if int(n) != len(exp) || err != nil {
			t.Errorf("formatter.WriteTo(builder) = (%d, %v), want (%d, nil)", n, err, len(exp))
		}
		out := b.String()
		testutils.DiffLine(t, strings.ToUpper(exp), strings.ToUpper(out))

	default:
		t.Errorf("result.Value.(type) = %T, want %T", seq, GenBank{})
	}
}
