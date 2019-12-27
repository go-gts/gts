package gts

import (
	"fmt"
	"io"

	pars "gopkg.in/ktnyt/pars.v2"
)

type Formatter interface {
	fmt.Stringer
	io.WriterTo
}

// RecordParser will attempt to parse a single sequence record.
var RecordParser = pars.Any(GenBankParser)

// RecordScanner will scan one sequence record at a time.
type RecordScanner struct {
	state  *pars.State
	result pars.Result
	err    error
}

func NewRecordScanner(r io.Reader) *RecordScanner {
	return &RecordScanner{pars.NewState(r), pars.Result{}, nil}
}

func peelError(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return err
	}
	return peelError(u.Unwrap())
}

func (s *RecordScanner) Scan() bool {
	s.result, s.err = RecordParser.Parse(s.state)
	return s.err == nil
}

func (s *RecordScanner) Record() Record {
	if rec, ok := s.result.Value.(Record); ok {
		return rec
	}
	return nil
}

func (s RecordScanner) Err() error {
	if peelError(s.err) == io.EOF {
		return nil
	}
	return s.err
}
