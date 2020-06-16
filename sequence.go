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

func insert(p []byte, pos int, q []byte) []byte {
	return append(p[:pos], append(q, p[pos:]...)...)
}

// Insert a sequence at the given position.
func Insert(host Sequence, pos int, guest Sequence) Sequence {
	var ff FeatureTable
	for _, f := range host.Features() {
		f.Location = f.Location.Shift(pos, Len(guest))
		ff = ff.Insert(f)
	}
	for _, f := range guest.Features() {
		f.Location = f.Location.Expand(0, pos)
		ff = ff.Insert(f)
	}
	p := insert(host.Bytes(), pos, guest.Bytes())
	return WithBytes(WithFeatures(host, ff), p)

}

// Embed a sequence at the given position.
func Embed(host Sequence, pos int, guest Sequence) Sequence {
	var ff FeatureTable
	for _, f := range host.Features() {
		f.Location = f.Location.Expand(pos, Len(guest))
		ff = ff.Insert(f)
	}
	for _, f := range guest.Features() {
		f.Location = f.Location.Expand(0, pos)
		ff = ff.Insert(f)
	}
	p := insert(host.Bytes(), pos, guest.Bytes())
	return WithBytes(WithFeatures(host, ff), p)
}

// Delete a region of the sequence at the given position and length.
func Delete(seq Sequence, i, n int) Sequence {
	ff := make([]Feature, len(seq.Features()))
	for i, f := range seq.Features() {
		ff[i].Key = f.Key
		ff[i].Location = f.Location.Expand(i, -n)
		ff[i].Qualifiers = f.Qualifiers
	}
	q := seq.Bytes()
	p := make([]byte, len(q)-n)
	copy(p[:i], q[:i])
	copy(p[i:], q[i+n:])
	return WithBytes(WithFeatures(seq, ff), p)
}

// Slice returns a subsequence of the given sequence starting at start and up
// to end. The target sequence region is copied.
func Slice(seq Sequence, start, end int) Sequence {
	p := make([]byte, end-start)
	copy(p, seq.Bytes()[start:end])
	var ff []Feature
	for _, f := range seq.Features() {
		loc := f.Location.Expand(start, -start).Expand(end-1, end-Len(seq))
		if !isBetween(loc) || isBetween(f.Location) {
			f.Location = loc
			ff = append(ff, f)
		}
	}
	return WithBytes(WithFeatures(seq, ff), p)
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
				f.Location = f.Location.Expand(0, len(p))
				ff = ff.Insert(f)
			}
			p = append(p, seq.Bytes()...)
		}
		return WithBytes(WithFeatures(head, ff), p)
	}
}

// Reverse returns a Sequence object with the byte representation in the
// reversed order.
func Reverse(seq Sequence) Sequence {
	var ff FeatureTable
	for _, f := range seq.Features() {
		ff = ff.Insert(Feature{f.Key, f.Location.Reverse(Len(seq)), f.Qualifiers, f.order})
	}
	p := make([]byte, Len(seq))
	copy(p, seq.Bytes())
	for l, r := 0, len(p)-1; l < r; l, r = l+1, r-1 {
		p[l], p[r] = p[r], p[l]
	}
	return WithBytes(WithFeatures(seq, ff), p)
}

// Rotate returns a Sequence object whose coordinates are shifted by the given
// amount.
func Rotate(seq Sequence, n int) Sequence {
	for Len(seq) > 0 && n < 0 {
		n += Len(seq)
	}
	var ff FeatureTable
	for _, f := range seq.Features() {
		loc := f.Location.Expand(0, n).Normalize(Len(seq))
		ff = ff.Insert(Feature{f.Key, loc, f.Qualifiers, f.order})
	}
	p := seq.Bytes()
	p = append(p[n:], p[:n]...)
	return WithBytes(WithFeatures(seq, ff), p)
}
