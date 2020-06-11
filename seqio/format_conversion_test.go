package seqio

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/testutils"
	"github.com/go-pars/pars"
)

func TestFormatConversion(t *testing.T) {
	in := testutils.ReadTestfile(t, "NC_001422.gb")
	out := testutils.ReadTestfile(t, "NC_001422.fasta")

	state := pars.FromString(in)
	parser := pars.AsParser(GenBankParser)

	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("parser returned %v\nBuffer:\n%q", err, string(result.Token))
		return
	}

	gb := result.Value.(GenBank)
	seq := gts.New(gb.Info(), gb.Features(), bytes.ToUpper(gb.Bytes()))

	f := NewFormatter(seq, FastaFile)
	builder := strings.Builder{}
	n, err := f.WriteTo(&builder)
	if err != nil {
		t.Errorf("f.WriteTo = %d, %v", n, err)
	}

	s := builder.String()
	testutils.Diff(t, s, out)
}
