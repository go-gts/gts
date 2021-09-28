package seqio

// Reference represents a reference of a record.
type Reference struct {
	Number  int
	Info    string
	Authors string
	Group   string
	Title   string
	Journal string
	Xref    map[string]string
	Comment string
}
