package gts

import (
	"io"
)

// DefaultWriter returns the default formatter for the given record.
func DefaultWriter(seq Sequence) io.WriterTo {
	if wt, ok := seq.(io.WriterTo); ok {
		return wt
	}
	if rec, ok := seq.(Record); ok {
		return GenBankWriter{rec}
	}
	return FastaWriter{seq}
}

// NewSequenceWriter returns a sequence writer for the given filetype.
func NewSequenceWriter(seq Sequence, filetype FileType) io.WriterTo {
	switch filetype {
	case DefaultFile:
		return DefaultWriter(seq)
	case JSONFile:
		return NewEncoderWriter(seq, NewJSONEncoder)
	case YAMLFile:
		return NewEncoderWriter(seq, NewYAMLEncoder)
	case MsgpackFile:
		return NewEncoderWriter(seq, NewMsgpackEncoder)
	case FastaFile:
		return FastaWriter{seq}
	case GenBankFile:
		return GenBankWriter{seq}
	default:
		return DefaultWriter(seq)
	}
}

// NewRecordWriter returns a record writer for the given filetype.
func NewRecordWriter(rec Record, filetype FileType) io.WriterTo {
	switch filetype {
	case DefaultFile:
		return DefaultWriter(rec)
	case GenBankFile:
		return GenBankWriter{rec}
	case JSONFile:
		return NewEncoderWriter(rec, NewJSONEncoder)
	case YAMLFile:
		return NewEncoderWriter(rec, NewYAMLEncoder)
	case MsgpackFile:
		return NewEncoderWriter(rec, NewMsgpackEncoder)
	default:
		return DefaultWriter(rec)
	}
}
