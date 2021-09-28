package seqio

import (
	"github.com/go-gts/gts"
)

type Callback func(seq gts.Sequence) error
type Manipulator func(i int, header interface{}, ff gts.Features) (Callback, error)

type IOStream interface {
	Peek() error
	ForEach(manip Manipulator) error

	PushHeader(header interface{}) error
	PushFeatures(ff gts.Features) error
	PushSequence(seq gts.Sequence) error
}
