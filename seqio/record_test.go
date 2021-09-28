package seqio

import (
	"errors"
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
)

var errRecordManipulate = errors.New("Record Manipulate error")
var errManipulatorCallback = errors.New("Manipulator Callback error")

var recordManipulateTests = []struct {
	manip Manipulator
	err   error
}{
	// case 1
	{
		func(i int, header interface{}, ff gts.Features) (Callback, error) {
			return func(seq gts.Sequence) error {
				return nil
			}, nil
		},
		nil,
	},

	// case 2
	{
		func(i int, header interface{}, ff gts.Features) (Callback, error) {
			return nil, errRecordManipulate
		},
		errRecordManipulate,
	},

	// case 3
	{
		func(i int, header interface{}, ff gts.Features) (Callback, error) {
			return func(seq gts.Sequence) error {
				return errManipulatorCallback
			}, nil
		},
		errManipulatorCallback,
	},

	// case 4
	{
		func(i int, header interface{}, ff gts.Features) (Callback, error) {
			return nil, nil
		},
		errNilManipulatorCallback,
	},
}

func TestRecordManipulate(t *testing.T) {
	rec := Record{}

	for i, tt := range recordManipulateTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			if err := rec.Manipulate(tt.manip, 0); err != tt.err {
				t.Errorf("rec.Manipulate(tt.manip, 0) = %v, want %v", err, tt.err)
			}
		})
	}
}
