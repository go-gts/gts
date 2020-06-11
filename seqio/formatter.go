package seqio

import (
	"io"

	"github.com/go-gts/gts"
)

// Formatter represents a formattable object.
type Formatter io.WriterTo

// NewFormatter returns a sequence formatter for the given FileType.
func NewFormatter(seq gts.Sequence, filetype FileType) Formatter {
	switch filetype {
	case GenBankFile:
		return GenBankFormatter{seq}
	case FastaFile:
		return FastaFormatter{seq, 70}
	default:
		return NewAutoFormatter(seq)
	}
}

// NewAutoFormatter returns a sequence formatter based on the sequence type.
func NewAutoFormatter(seq gts.Sequence) Formatter {
	switch seq.(type) {
	case GenBank:
		return GenBankFormatter{seq}
	case Fasta:
		return FastaFormatter{seq, 70}
	}
	switch seq.Info().(type) {
	case GenBankFields:
		return GenBankFormatter{seq}
	case string:
		return FastaFormatter{seq, 70}
	}
	return FastaFormatter{seq, 70}
}
