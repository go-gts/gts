package gts

import (
	"testing"
)

func TestComplement(t *testing.T) {
	in := Seq("atgcATGCnN")
	out := Complement(in)
	e := Seq("tacgTACGnN")
	if !Equal(out, e) {
		t.Errorf("Complement(%q) = %q, want %q", in, out, e)
	}
}
