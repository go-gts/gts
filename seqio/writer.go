package seqio

import (
	"fmt"
	"io"

	"github.com/go-gts/gts"
)

type SeqWriter interface {
	WriteSeq(seq gts.Sequence) (int, error)
}

type AutoWriter struct {
	w  io.Writer
	sw SeqWriter
}

func NewWriter(w io.Writer, filetype FileType) SeqWriter {
	switch filetype {
	case FastaFile:
		return FastaWriter{w}
	case GenBankFile:
		return GenBankWriter{w}
	default:
		return AutoWriter{w, nil}
	}
}

func detectWriter(seq gts.Sequence, w io.Writer) (SeqWriter, error) {
	switch seq.(type) {
	case GenBank, *GenBank:
		return GenBankWriter{w}, nil
	case Fasta, *Fasta:
		return FastaWriter{w}, nil
	default:
		switch info := seq.Info().(type) {
		case GenBankFields:
			return GenBankWriter{w}, nil
		case string, fmt.Stringer:
			return FastaWriter{w}, nil
		default:
			return nil, fmt.Errorf("gts does not know how to format a sequence with metadata type `%T`", info)
		}
	}
}

func (w AutoWriter) WriteSeq(seq gts.Sequence) (int, error) {
	if w.sw == nil {
		sw, err := detectWriter(seq, w.w)
		if err != nil {
			return 0, err
		}
		w.sw = sw
	}
	return w.sw.WriteSeq(seq)
}
