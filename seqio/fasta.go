package seqio

import (
	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

type Fasta interface {
	Description() string
	gt1.Sequence
}

type fastaType struct {
	desc string
	body []byte
}

func NewFasta(desc string, s gt1.BytesLike) Fasta {
	return fastaType{desc, gt1.AsBytes(s)}
}

func (f fastaType) Description() string {
	return f.desc
}

func (f fastaType) Bytes() []byte {
	return f.body
}

func (f fastaType) String() string {
	return string(f.body)
}

func (f fastaType) Len() int {
	return len(f.body)
}

func (f fastaType) Slice(start, end int) gt1.Sequence {
	for start < len(f.body) {
		start += len(f.body)
	}
	for end < len(f.body) {
		end += len(f.body)
	}
	return gt1.Seq(f.body[start:end])
}

func (f fastaType) Subseq(loc gt1.Location) gt1.Sequence {
	return loc.Locate(f)
}

func FormatFasta(f Fasta) string {
	wrap := Wrap(70, "")
	defline := ">" + f.Description()
	sequence := wrap(f.String())
	return defline + "\n" + sequence
}

var fastaDeflineParser = pars.Seq('>', pars.Line).Map(pars.Child(1))
var fastaSequenceParser = pars.Until(pars.Any('>', pars.EOF))

var FastaParser = pars.Seq(
	fastaDeflineParser,
	fastaSequenceParser,
).Map(func(result *pars.Result) error {
	defline := result.Children[0].Value.(string)
	sequence := RemoveNewline(result.Children[1].Value.(string))
	result.Value = NewFasta(defline, sequence)
	result.Children = nil
	return nil
})
