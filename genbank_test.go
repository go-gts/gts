package gts

import (
	"testing"

	pars "gopkg.in/ktnyt/pars.v2"
)

func TestGenBankIO(t *testing.T) {
	in := ReadGolden(t)
	state := pars.FromString(in)
	parser := pars.AsParser(GenBankParser)
	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("parser returned %v\nBuffer:\n%q", err, string(result.Token))
	}

	switch rec := result.Value.(type) {
	case Record:
	default:
		t.Errorf("result.Value.(type) = %T, want %T", rec, Record(GenBank{}))
	}
}
