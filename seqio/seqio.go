package seqio

import (
	"errors"
	"io"

	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
)

type SequenceHandler func(seq gts.Sequence) error
type FeatureHandler func(i int, header interface{}, ff gts.Features) (SequenceHandler, error)

type IStream interface {
	Peek() error
	ForEach(fh FeatureHandler) error
}

type OStream interface {
	PushHeader(header interface{}) error
	PushFeatures(ff gts.Features) error
	PushSequence(seq gts.Sequence) error
}

type IOStream interface {
	IStream
	OStream
}

var IStreamGenerators = []func(r io.Reader, w io.Writer) (IStream, OStream){
	func(r io.Reader, w io.Writer) (IStream, OStream) {
		return NewGenBankIStream(r), NewGenBankOstream(w)
	},
}

func NewSeqIO(r io.Reader, w io.Writer) (IStream, OStream, error) {
	state := pars.NewState(r)
	for _, gen := range IStreamGenerators {
		istream, ostream := gen(state, w)
		if istream.Peek() == nil {
			return istream, ostream, nil
		}
	}
	return nil, nil, errors.New("gq cannot process the given input stream")
}
