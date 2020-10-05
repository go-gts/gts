package gts

import "testing"

var moleculeCounterTests = []struct {
	in  Molecule
	out string
}{
	{DNA, "bases"},
	{RNA, "bases"},
	{AA, "residues"},
	{SingleStrandDNA, "bases"},
}

func TestMoleculeCounter(t *testing.T) {
	for _, tt := range moleculeCounterTests {
		out := tt.in.Counter()
		if out != tt.out {
			t.Errorf("%q.Counter = %q, want %q", tt.in, out, tt.out)
		}
	}
}

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
