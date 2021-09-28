package seqio

import (
	"net/url"

	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
)

// GFF3GenomeBuild represents the genome build directive.
type GFF3GenomeBuild struct {
	Source string
	Name   string
}

// GFF3Header represents the directives of a GFF3 record other than the
// features and sequence.
type GFF3Header struct {
	Version string
	ID      string
	Region  gts.Segment

	FeatureOntology   *url.URL
	AttributeOntology *url.URL
	SourceOntology    *url.URL

	Species     *url.URL
	GenomeBuild GFF3GenomeBuild
}

var gff3VersionParser = pars.Seq("##gff-version ", pars.Line)
