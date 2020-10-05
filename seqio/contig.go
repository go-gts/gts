package seqio

import (
	"fmt"

	"github.com/go-gts/gts"
)

// Contig represents a contig field.
type Contig struct {
	Accession string
	Region    gts.Segment
}

// String satisfies the fmt.Stringer interface.
func (contig Contig) String() string {
	if contig.Accession == "" {
		return ""
	}
	head, tail := gts.Unpack(contig.Region)
	return fmt.Sprintf("join(%s:%d..%d)", contig.Accession, head+1, tail)
}
