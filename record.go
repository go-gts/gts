package gts

// Record is the interface for sequence records with metadata and features.
type Record interface {
	Metadata() interface{}
	Features() FeatureTable
	Sequence
}
