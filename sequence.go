package gts

import (
	"bytes"
	"reflect"
)

// Molecule represents the sequence molecule type.
type Molecule string

// Molecule constants for DNA, RNA, and amino acid (AA).
const (
	DNA Molecule = "DNA"
	RNA Molecule = "RNA"
	AA  Molecule = "AA"
)

// Sequence represents a biological sequence. All sequences are expected to be
// able to return its metadata and byte representation.
type Sequence interface {
	Info() interface{}
	Features() FeatureTable
	Bytes() []byte
}

// Len returns the length of the given Sequence.
func Len(seq Sequence) int {
	return len(seq.Bytes())
}

// Equal tests if the given sequences are identical.
func Equal(a, b Sequence) bool {
	return reflect.DeepEqual(a.Info(), b.Info()) && bytes.Equal(a.Bytes(), b.Bytes())
}

// BasicSequence represents the most basic Sequence object.
type BasicSequence struct {
	info  interface{}
	table FeatureTable
	data  []byte
}

// New returns a new GTS object with the given metadata and bytes.
func New(info interface{}, table FeatureTable, p []byte) BasicSequence {
	return BasicSequence{info, table, p}
}

// Info returns the metadata of the sequence.
func (seq BasicSequence) Info() interface{} {
	return seq.info
}

// Features returns the feature table of the sequence.
func (seq BasicSequence) Features() FeatureTable {
	return seq.table
}

// Bytes returns the byte representation of the sequence.
func (seq BasicSequence) Bytes() []byte {
	return seq.data
}

// Slice returns a subsequence of the given sequence starting at start and up
// to end. The target sequence region is copied.
func Slice(seq Sequence, start, end int) BasicSequence {
	p := make([]byte, end-start)
	copy(p, seq.Bytes()[start:end])
	return New(seq.Info(), seq.Features(), p)
}

// Reverse returns a Sequence object with the byte representation in the
// reversed order.
func Reverse(seq Sequence) Sequence {
	p := make([]byte, Len(seq))
	copy(p, seq.Bytes())
	for l, r := 0, len(p)-1; l < r; l, r = l+1, r-1 {
		p[l], p[r] = p[r], p[l]
	}
	return New(seq.Info(), seq.Features(), p)
}
