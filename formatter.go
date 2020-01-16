package gts

import (
	"bytes"
	"io"
)

// Formatter can be formatted to a string or write to an io.Writer.
type Formatter interface {
	io.WriterTo
}

// EncoderFormatter provides an interface for encoding objects with the given
// Encoder.
type EncoderFormatter struct {
	value interface{}
	ctor  EncoderConstructor
}

// NewEncoderFormatter creates a new EncoderFormatter.
func NewEncoderFormatter(v interface{}, ctor EncoderConstructor) EncoderFormatter {
	return EncoderFormatter{v, ctor}
}

// WriteTo satisfies the WriterTo interface.
func (f EncoderFormatter) WriteTo(w io.Writer) (int64, error) {
	b := &bytes.Buffer{}
	enc := f.ctor(b)
	if err := enc.Encode(f.value); err != nil {
		return 0, err
	}
	n, err := w.Write(b.Bytes())
	return int64(n), err
}
