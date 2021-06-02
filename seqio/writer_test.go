package seqio

import (
	"strings"
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
)

var writerTests = []struct {
	filename string
	filetype FileType
}{
	{"NC_001422.fasta", FastaFile},
	{"NC_001422.fasta", DefaultFile},
	{"NC_001422.gb", GenBankFile},
	{"NC_001422.gb", DefaultFile},
}

func TestWriter(t *testing.T) {
	for _, tt := range writerTests {
		in := testutils.ReadTestfile(t, tt.filename)
		scanner := NewAutoScanner(strings.NewReader(in))
		if !scanner.Scan() {
			t.Errorf("failed to scan test file %s", tt.filename)
		}

		seq := scanner.Value()

		w := &strings.Builder{}
		n, err := NewWriter(w, tt.filetype).WriteSeq(seq)
		if n != len(in) || err != nil {
			t.Errorf("writer.WriteSeq(seq) = (%d, %v), want (%d, nil)", n, err, len(in))
		}
		testutils.DiffLine(t, w.String(), in)

		seq = gts.New(seq.Info(), seq.Features(), seq.Bytes())

		w.Reset()
		n, err = NewWriter(w, tt.filetype).WriteSeq(seq)
		if n != len(in) || err != nil {
			t.Errorf("writer.WriteSeq(seq) = (%d, %v), want (%d, nil)", n, err, len(in))
		}
		testutils.DiffLine(t, w.String(), in)
	}
}

func TestWriterFail(t *testing.T) {
	w := &strings.Builder{}
	n, err := NewWriter(w, DefaultFile).WriteSeq(gts.New(nil, nil, nil))
	if n != 0 || err == nil {
		t.Errorf("writer.WriteSeq(seq) = (%d, nil), want (0, error)", n)
	}
}
