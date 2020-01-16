package gts

// Range represents a start and end index pair.
type Range struct {
	Start int
	End   int
}

// Reference represents a reference of a record.
type Reference struct {
	Number  int               `json:"number" yaml:"number" msgpack:"number"`
	Ranges  []Range           `json:"ranges,omitempty" yaml:"ranges,omitempty" msgpack:"ranges,omitempty"`
	Authors string            `json:"authors" yaml:"authors" msgpack:"authors"`
	Group   string            `json:"group" yaml:"group" msgpack:"group"`
	Title   string            `json:"title" yaml:"title" msgpack:"title"`
	Journal string            `json:"journal" yaml:"journal" msgpack:"journal"`
	Xref    map[string]string `json:"xref,omitempty" yaml:"xref,omitempty" msgpack:"xref,omitempty"`
	Comment string            `json:"comment" yaml:"comment" msgpack:"comment"`
}
