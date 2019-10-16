package seqio

import (
	"io"

	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

type parserDef struct {
	Name string
	Func pars.Parser
}

var parserDefs = []parserDef{
	parserDef{"genbank", GenBankParser},
	parserDef{"fasta", FastaParser},
}

type Scanner struct {
	state    *pars.State
	filetype string
	sequence gt1.Sequence
	done     bool
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{pars.NewState(r), "", nil, false}
}

func (scanner *Scanner) Scan() bool {
	if scanner.done {
		return false
	}

	scanner.sequence = nil

	for _, def := range parserDefs {
		scanner.state.Mark()
		result := &pars.Result{}

		err := def.Func(scanner.state, result)

		if err == io.EOF {
			scanner.done = true
			err = nil
		}

		if err == nil {
			scanner.state.Unmark()

			if seq, ok := result.Value.(gt1.Sequence); ok {
				scanner.sequence = seq
				return true
			}

			panic("encountered unexpected result type in Scan")
		}

		scanner.state.Jump()
	}

	scanner.state.Jump()
	scanner.done = true

	return false
}

func (scanner Scanner) Seq() gt1.Sequence {
	return scanner.sequence
}
