package gts

import "testing"

var asMoleculeTests = []Molecule{
	DNA,
	RNA,
	AA,
	SingleStrandDNA,
}

func TestAsMolecule(t *testing.T) {
	for _, in := range asMoleculeTests {
		out, err := AsMolecule(string(in))
		if err != nil {
			t.Errorf("AsMolecule(%q): %v", string(in), err)
		}
		if out != in {
			t.Errorf("AsMolecule(%q) = %q, expected %q", string(in), out, string(in))
		}
	}

	_, err := AsMolecule("")
	if err == nil {
		t.Errorf("expected error in AsMolecule(%q)", "")
	}
}
