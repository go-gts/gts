package seqio

import (
	"io"

	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
)

var sequenceParsers = []pars.Parser{
	GenBankParser,
	FastaParser,
}

// Scanner represents a sequence file scanner.
type Scanner struct {
	p   pars.Parser
	s   *pars.State
	res pars.Result
	err error
}

// NewScanner creates a new sequence scanner.
func NewScanner(p pars.Parser, r io.Reader) *Scanner {
	return &Scanner{p, pars.NewState(r), pars.Result{}, nil}
}

// NewAutoScanner creates a new sequence scanner which will automatically
// detect the sequence format from a list of known parsers on the first scan.
func NewAutoScanner(r io.Reader) *Scanner {
	return NewScanner(nil, r)
}

func dig(err error) error {
	if v, ok := err.(interface{ Unwrap() error }); ok {
		return dig(v.Unwrap())
	}
	return err
}

// Scan advances the scanner using the given parser. If the parser is not yet
// specified, the first scan will match one of the known parsers.
func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	if s.p == nil {
		errs := make([]struct {
			err error
			pos pars.Position
		}, len(sequenceParsers))
		for i, p := range sequenceParsers {
			s.s.Push()
			s.res, errs[i].err = p.Parse(s.s)
			if errs[i].err == nil {
				s.s.Drop()
				s.p = p
				return true
			}
			errs[i].pos = s.s.Position()
			s.s.Pop()
		}
		argmax := 0
		maxpos := pars.Position{Line: 0, Byte: 0}
		for i, v := range errs {
			if maxpos.Less(v.pos) {
				argmax = i
				maxpos = v.pos
			}
		}
		s.err = errs[argmax].err
		return false
	}

	s.res, s.err = s.p.Parse(s.s)
	return s.err == nil || dig(s.err) == io.EOF
}

// Value returns the most recently scanned sequence value.
func (s Scanner) Value() gts.Sequence {
	return s.res.Value.(gts.Sequence)
}

// Err returns the first non-EOF error that was encountered by the scanner.
func (s Scanner) Err() error {
	if s.err == nil || dig(s.err) == io.EOF {
		return nil
	}
	return s.err
}
