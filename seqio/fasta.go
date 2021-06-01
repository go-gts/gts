package seqio

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
	"github.com/go-wrap/wrap"
)

// Fasta represents a FASTA format sequence object.
type Fasta struct {
	Desc string
	Data []byte
}

// Info returns the metadata of the sequence.
func (f Fasta) Info() interface{} {
	return f.Desc
}

// Features returns the feature table of the sequence.
func (f Fasta) Features() gts.FeatureSlice {
	return nil
}

// Bytes returns the byte representation of the sequence.
func (f Fasta) Bytes() []byte {
	return f.Data
}

// WriteTo satisfies the io.WriterTo interface.
func (f Fasta) WriteTo(w io.Writer) (int64, error) {
	desc := strings.ReplaceAll(f.Desc, "\n", " ")
	data := wrap.Force(string(f.Data), 70)
	s := fmt.Sprintf(">%s\n%s\n", desc, data)
	n, err := io.WriteString(w, s)
	return int64(n), err
}

// FastaWriter writes a gts.Sequence to an io.Writer in FASTA format.
type FastaWriter struct {
	w io.Writer
}

// WriteSeq satisfies the seqio.SeqWriter interface.
func (w FastaWriter) WriteSeq(seq gts.Sequence) (int, error) {
	switch v := seq.(type) {
	case Fasta:
		n, err := v.WriteTo(w.w)
		return int(n), err
	case *Fasta:
		return w.WriteSeq(*v)
	default:
		switch info := v.Info().(type) {
		case string:
			f := Fasta{info, v.Bytes()}
			return w.WriteSeq(f)
		case fmt.Stringer:
			f := Fasta{info.String(), v.Bytes()}
			return w.WriteSeq(f)
		default:
			return 0, fmt.Errorf("gts does not know how to format a sequence with metadata type `%T` as FASTA", info)
		}
	}
}

// FastaParser attempts to parse a single FASTA file entry.
var FastaParser = pars.Seq(
	'>', pars.Line, pars.Until(pars.Any('>', pars.End)),
).Map(func(result *pars.Result) error {
	desc := string(result.Children[1].Token)
	body := result.Children[2].Token
	lines := bytes.Split(body, []byte{'\n'})
	data := bytes.Join(lines, nil)
	result.SetValue(Fasta{desc, data})
	return nil
})
