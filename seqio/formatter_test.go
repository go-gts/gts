package seqio

import (
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
)

var basicSeq = gts.New(nil, nil, nil)
var fastaSeq = Fasta{"", nil}
var genbankSeq = NewGenBank(GenBankFields{}, nil, nil)
var strInfoSeq = gts.New("", nil, nil)
var gbInfoSeq = gts.New(GenBankFields{}, nil, nil)

var newFormatterTests = []struct {
	in  gts.Sequence
	ft  FileType
	out Formatter
}{
	{basicSeq, FastaFile, FastaFormatter{basicSeq, 70}},
	{basicSeq, GenBankFile, GenBankFormatter{basicSeq}},
	{basicSeq, DefaultFile, FastaFormatter{basicSeq, 70}},

	{fastaSeq, DefaultFile, FastaFormatter{fastaSeq, 70}},
	{genbankSeq, DefaultFile, GenBankFormatter{genbankSeq}},

	{strInfoSeq, DefaultFile, FastaFormatter{fastaSeq, 70}},
	{gbInfoSeq, DefaultFile, GenBankFormatter{genbankSeq}},
}

func TestNewFormatter(t *testing.T) {
	for _, tt := range newFormatterTests {
		out := NewFormatter(tt.in, tt.ft)
		testutils.Equals(t, out, tt.out)
	}
}
