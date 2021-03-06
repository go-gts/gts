package seqio

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-pars/pars"
)

func TestFastaIO(t *testing.T) {
	in := testutils.ReadTestfile(t, "NC_001422.fasta")
	state := pars.FromString(in)
	parser := pars.AsParser(FastaParser)

	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("parser returned %v\nBuffer:\n%q", err, string(result.Token))
	}

	switch seq := result.Value.(type) {
	case Fasta:
		if gts.Len(seq) != 5386 {
			t.Errorf("gts.Len(seq) = %d, want 5386", gts.Len(seq))
		}
		if seq.Info() == nil {
			t.Error("seq.Info() is nil")
		}
		if seq.Features() != nil {
			t.Error("seq.Features() is not nil")
		}
		t.Run("format from *Fasta", func(t *testing.T) {
			b := strings.Builder{}
			f := FastaFormatter{&seq, 70}
			n, err := f.WriteTo(&b)
			if int(n) != len([]byte(in)) || err != nil {
				t.Errorf("f.WriteTo(&b) = (%d, %v), want %d, nil", n, err, len(in))
				return
			}
			out := b.String()
			testutils.DiffLine(t, in, out)
		})
		t.Run("format from BasicSequence", func(t *testing.T) {
			b := strings.Builder{}
			f := FastaFormatter{gts.Copy(seq), 70}
			n, err := f.WriteTo(&b)
			if int(n) != len([]byte(in)) || err != nil {
				t.Errorf("f.WriteTo(&b) = (%d, %v), want %d, nil", n, err, len(in))
				return
			}
			out := b.String()
			testutils.DiffLine(t, in, out)
		})
	default:
		t.Errorf("result.Value.(type) = %T, want %T", seq, Fasta{})
	}
}

func TestFastaIOFail(t *testing.T) {
	w := bytes.Buffer{}
	n, err := FastaFormatter{gts.New(nil, nil, nil), 70}.WriteTo(&w)
	if n != 0 || err == nil {
		t.Errorf("formatting an empty Sequence should return an error")
	}
}
