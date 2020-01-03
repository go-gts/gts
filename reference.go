package gts

// Range represents a start and end index pair.
type Range struct {
	Start int
	End   int
}

// Reference represents a reference of a record.
type Reference struct {
	Number  int
	Ranges  []Range
	Authors string
	Group   string
	Title   string
	Journal string
	Xref    map[string]string
	Comment string
}
