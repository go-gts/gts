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
	if v, ok := seq.(interface {
		Len() int
	}); ok {
		return v.Len()

	}
	return len(seq.Bytes())
}

// Equal tests if the given sequences are identical.
func Equal(a, b Sequence) bool {
	return reflect.DeepEqual(a.Info(), b.Info()) &&
		reflect.DeepEqual(a.Features(), b.Features()) &&
		bytes.Equal(a.Bytes(), b.Bytes())
}

// BasicSequence represents the most basic Sequence object.
type BasicSequence struct {
	info  interface{}
	table FeatureTable
	data  []byte
}

// New returns a new Sequence object with the given values.
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

// Copy returns a shallow copy of the given sequence.
func Copy(seq Sequence) BasicSequence {
	return New(seq.Info(), seq.Features(), seq.Bytes())
}

type withInterface interface {
	WithInfo(info interface{}) Sequence
	WithFeatures(ff FeatureTable) Sequence
	WithBytes(p []byte) Sequence
}

// WithInfo creates a shallow copy of the given Sequence object and swaps the
// metadata with the given value.
func WithInfo(seq Sequence, info interface{}) Sequence {
	switch v := seq.(type) {
	case withInterface:
		return v.WithInfo(info)
	default:
		return New(info, seq.Features(), seq.Bytes())
	}
}

// WithFeatures creates a shallow copy of the given Sequence object and swaps
// the feature table with the given features.
func WithFeatures(seq Sequence, ff []Feature) Sequence {
	switch v := seq.(type) {
	case withInterface:
		return v.WithFeatures(ff)
	default:
		return New(seq.Info(), ff, seq.Bytes())
	}
}

// WithBytes creates a shallow copy of the given Sequence object and swaps the
// byte representation with the given byte slice.
func WithBytes(seq Sequence, p []byte) Sequence {
	switch v := seq.(type) {
	case withInterface:
		return v.WithBytes(p)
	default:
		return New(seq.Info(), seq.Features(), p)
	}
}

// Slice returns a subsequence of the given sequence starting at start and up
// to end. The target sequence region is copied.
func Slice(seq Sequence, start, end int) BasicSequence {
	p := make([]byte, end-start)
	copy(p, seq.Bytes()[start:end])
	before, after := Span{0, -start}, Span{end, end - Len(seq)}
	ff := make([]Feature, 0)
	for _, f := range seq.Features() {
		loc := f.Location.Shift(before).Shift(after)
		if !isBetween(loc) || isBetween(f.Location) {
			f.Location = loc
			ff = append(ff, f)
		}
	}
	return New(seq.Info(), seq.Features(), p)
}

// Concat takes the given Sequences and concatenates them into a single
// Sequence.
func Concat(ss ...Sequence) Sequence {
	switch len(ss) {
	case 0:
		return New(nil, nil, nil)
	case 1:
		return ss[0]
	default:
		head, tail := ss[0], ss[1:]
		ff, p := head.Features(), head.Bytes()
		for _, seq := range tail {
			for _, f := range seq.Features() {
				f.Location = f.Location.Shift(Span{0, len(p)})
				ff = ff.Insert(f)
			}
			p = append(p, seq.Bytes()...)
		}
		head = WithFeatures(head, ff)
		head = WithBytes(head, p)
		return head
	}
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
