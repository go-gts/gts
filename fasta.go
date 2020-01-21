package gts

import (
	"fmt"
	"io"
	"strings"
)

// Fasta represents a FASTA sequence object.
type Fasta struct {
	Desc string
	Body []byte
}

// Info returns the metadata of the sequence.
func (f Fasta) Info() interface{} {
	return f.Desc
}

// Bytes returns the byte representation of the sequence.
func (f Fasta) Bytes() []byte {
	return f.Body
}

// Insert a sequence at the specified position.
func (f *Fasta) Insert(pos int, arg Sequence) error {
	return (*mutableByteSlice)(&(f.Body)).Insert(pos, arg)
}

// Delete given number of bases from the specified position.
func (f *Fasta) Delete(pos, arg int) error {
	return (*mutableByteSlice)(&(f.Body)).Delete(pos, arg)
}

// Replace the bases from the specified position with the given sequence.
func (f *Fasta) Replace(pos int, arg Sequence) error {
	return (*mutableByteSlice)(&(f.Body)).Replace(pos, arg)
}

// String satisfies the fmt.Stringer interface.
func (f Fasta) String() string {
	b := strings.Builder{}
	b.WriteByte('>')
	i := strings.IndexByte(f.Desc, '\n')
	if i < 0 {
		i = len(f.Desc)
	}
	b.WriteString(f.Desc[:i])
	b.WriteByte('\n')
	for i := 0; i < len(f.Body); i += 80 {
		j := min(i+80, len(f.Body))
		b.Write(f.Body[i:j])
		b.WriteByte('\n')
	}
	return b.String()
}

// WriteTo satisfies the io.WriterTo interface.
func (f Fasta) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, f.String())
	return int64(n), err
}

// FastaWriter attempts to format a sequence in FASTA format.
type FastaWriter struct {
	seq Sequence
}

// WriteTo satisfies the io.WriterTo interface.
func (ff FastaWriter) WriteTo(w io.Writer) (int64, error) {
	switch seq := ff.seq.(type) {
	case Fasta:
		return seq.WriteTo(w)
	case *Fasta:
		return seq.WriteTo(w)
	default:
		data := seq.Bytes()
		switch info := seq.Info().(type) {
		case string:
			return Fasta{info, data}.WriteTo(w)
		case fmt.Stringer:
			return Fasta{info.String(), data}.WriteTo(w)
		case GenBankFields:
			return Fasta{info.Version + " " + info.Definition, data}.WriteTo(w)
		default:
			return Fasta{"sequence", data}.WriteTo(w)
		}
	}
}
