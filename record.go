package gd

import "time"

type Organism struct {
	Species string
	Name    string
	Taxon   []string
}

type Reference struct {
	Number  int
	Start   int
	End     int
	Authors string
	Group   string
	Title   string
	Journal string
	Xref    map[string]string
	Comment string
}

type Feature struct {
	Key        string
	Location   Locator
	Properties PairList
}

type Record struct {
	LocusName  string
	Accessions []string
	Topology   string
	Version    string
	Molecule   string
	Class      string
	Division   string
	Dates      []time.Time
	DBLink     PairList

	Definition string
	Keywords   []string

	Source     Organism
	References []Reference
	Comment    string
	Features   []Feature
	Origin     Sequence
}

func (r Record) Length() int {
	return r.Origin.Length()
}
