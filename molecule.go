package gts

import "fmt"

// Molecule represents the sequence molecule type.
type Molecule string

// Counter returns the count word for the Molecule.
func (mol Molecule) Counter() string {
	switch mol {
	case AA:
		return "residues"
	default:
		return "bases"
	}
}

// Molecule constants for DNA, RNA, and amino acid (AA).
const (
	DNA Molecule = "DNA"
	RNA          = "RNA"
	AA           = "AA"

	SingleStrandDNA = "ss-DNA"
	DoubleStrandDNA = "ds-DNA"
)

// AsMolecule attempts to convert a string into a Molecule object.
func AsMolecule(s string) (Molecule, error) {
	switch s {
	case "DNA":
		return DNA, nil
	case "RNA":
		return RNA, nil
	case "AA":
		return AA, nil
	case "ss-DNA":
		return SingleStrandDNA, nil
	case "ds-DNA":
		return DoubleStrandDNA, nil
	}
	return "", fmt.Errorf("molecule type for %q not known", s)
}
