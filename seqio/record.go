package seqio

import (
	"errors"

	"github.com/go-gts/gts"
)

var errNilManipulatorCallback = errors.New("nil Manipulator Callback")

type Record struct {
	Header   interface{}
	Features gts.Features
	Sequence gts.Sequence
}

func (rec Record) Manipulate(manip FeatureHandler, i int) error {
	cb, err := manip(i, rec.Header, rec.Features)
	if err != nil {
		return err
	}
	if cb == nil {
		return errNilManipulatorCallback
	}
	return cb(rec.Sequence)
}
