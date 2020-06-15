package gts

// ReferenceRange represents a start and end index pair.
type ReferenceRange struct {
	Start int
	End   int
}

// Reference represents a reference of a record.
type Reference struct {
	Number  int
	Ranges  []ReferenceRange
	Authors string
	Group   string
	Title   string
	Journal string
	Xref    map[string]string
	Comment string
}
