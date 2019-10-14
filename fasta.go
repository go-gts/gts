package gt1

import (
	"github.com/ktnyt/pars"
)

type Fasta interface {
	Description() string
	Sequence
}

type fastaType struct {
	desc string
	body []byte
}

func NewFasta(desc string, s BytesLike) Fasta {
	return fastaType{desc, asBytes(s)}
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

func (f fastaType) Length() int {
	return len(f.body)
}

func (f fastaType) Slice(start, end int) Sequence {
	for start < len(f.body) {
		start += len(f.body)
	}
	for end < len(f.body) {
		end += len(f.body)
	}
	return Seq(f.body[start:end])
}

func (f fastaType) Subseq(loc Location) Sequence {
	return loc.Locate(f)
}

func FormatFasta(f Fasta) string {
	defline := f.Description()
	sequence := wrap(f.String(), 0, 70)
	return defline + "\n" + sequence
}

var fastaDeflineParser = pars.Seq('>', pars.Line).Map(pars.Child(1))
var fastaSequenceParser = pars.Until(pars.Any('>', pars.EOF))

var FastaParser = pars.Seq(
	fastaDeflineParser,
	fastaSequenceParser,
).Map(func(result *pars.Result) error {
	result.Value = NewFasta(
		result.Children[0].Value.(string),
		result.Children[1].Value.(string),
	)
	result.Children = nil
	return nil
})
