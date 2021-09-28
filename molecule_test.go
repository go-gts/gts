package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

var moleculeCounterTests = []struct {
	in  Molecule
	out string
}{
	{DNA, "bases"},             // case 1
	{RNA, "bases"},             // case 2
	{AA, "residues"},           // case 3
	{SingleStrandDNA, "bases"}, // case 4
	{DoubleStrandDNA, "bases"}, // case 5
}

func TestMoleculeCounter(t *testing.T) {
	for i, tt := range moleculeCounterTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := tt.in.Counter()
			if out != tt.out {
				t.Errorf("%q.Counter = %q, want %q", tt.in, out, tt.out)
			}
		})
	}
}

var asMoleculeTests = []Molecule{
	DNA,             // case 1
	RNA,             // case 2
	AA,              // case 3
	SingleStrandDNA, // case 4
	DoubleStrandDNA, // case 5
}

func TestAsMolecule(t *testing.T) {
	for i, in := range asMoleculeTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out, err := AsMolecule(string(in))
			if err != nil {
				t.Errorf("AsMolecule(%q): %v", string(in), err)
			}
			if out != in {
				t.Errorf("AsMolecule(%q) = %q, expected %q", string(in), out, string(in))
			}
		})
	}

	_, err := AsMolecule("")
	if err == nil {
		t.Errorf("expected error in AsMolecule(%q)", "")
	}
}
