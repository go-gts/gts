package gts

import "fmt"

// Molecule represents the sequence molecule type.
type Molecule string

// Molecule constants for DNA, RNA, and amino acid (AA).
const (
	DNA Molecule = "DNA"
	RNA          = "RNA"
	AA           = "AA"

	SingleStrandDNA = "ss-DNA"
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
	}
	return "", fmt.Errorf("molecule type for %q not known", s)
}
