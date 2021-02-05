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
func (f Fasta) Features() gts.FeatureTable {
	return nil
}

// Bytes returns the byte representation of the sequence.
func (f Fasta) Bytes() []byte {
	return f.Data
}

// FastaFormatter implements the Formatter interface for FASTA files.
type FastaFormatter struct {
	seq  gts.Sequence
	wrap int
}

// WriteTo satisfies the io.WriterTo interface.
func (ff FastaFormatter) WriteTo(w io.Writer) (int64, error) {
	switch seq := ff.seq.(type) {
	case Fasta:
		desc := strings.ReplaceAll(seq.Desc, "\n", " ")
		data := wrap.Force(string(seq.Data), ff.wrap)
		s := fmt.Sprintf(">%s\n%s\n", desc, data)
		n, err := io.WriteString(w, s)
		return int64(n), err
	case *Fasta:
		return FastaFormatter{*seq, ff.wrap}.WriteTo(w)
	default:
		switch info := seq.Info().(type) {
		case string:
			f := Fasta{info, seq.Bytes()}
			return FastaFormatter{f, ff.wrap}.WriteTo(w)
		case fmt.Stringer:
			f := Fasta{info.String(), seq.Bytes()}
			return FastaFormatter{f, ff.wrap}.WriteTo(w)
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
