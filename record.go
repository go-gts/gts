package gts

import "fmt"

// Record is the interface for sequence records with metadata and features.
type Record interface {
	Metadata() interface{}
	FeatureTable
	MutableSequence
}

// NewRecord creates a new record/
func NewRecord(meta interface{}, ff []Feature, p []byte) Record {
	seq := Seq(p)
	switch v := meta.(type) {
	case GenBankFields:
		return GenBank{v, FeatureList(ff), NewSequenceServer(seq)}
	default:
		err := fmt.Errorf("gts does not know how to create a record using metadata of type `%T`", v)
		panic(err)
	}
}

// DefaultFormatter returns the default formatter for the given record.
func DefaultFormatter(rec Record) Formatter {
	switch rec.Metadata().(type) {
	case GenBankFields:
		return GenBankFormatter{rec}
	default:
		return GenBankFormatter{rec}
	}
}
