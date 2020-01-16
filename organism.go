package gts

// Organism represents an organism of a record.
type Organism struct {
	Species string   `json:"species" yaml:"species" msgpack:"species"`
	Name    string   `json:"name" yaml:"name" msgpack:"name"`
	Taxon   []string `json:"taxon,omitempty" yaml:"taxon,omitempty" msgpack:"taxon,omitempty"`
}
