package gts

import (
	"bytes"
	"reflect"

	"github.com/go-flip/flip"
)

// Sequence represents a biological sequence. All sequences are expected to be
// able to return its metadata, associated features, and byte representation.
type Sequence interface {
	Info() interface{}
	Features() FeatureTable
	Bytes() []byte
}

// Len returns the length of the given Sequence by computing the length of the
// byte representation. If the sequece implements the `Len() int` method, the
// method will be called instead.
func Len(seq Sequence) int {
	if v, ok := seq.(interface {
		Len() int
	}); ok {
		return v.Len()
	}
	return len(seq.Bytes())
}

// Equal tests if the given sequences are identical by comparing the deep
// equality of the metadata, features, and byte representations.
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

type withInfo interface {
	WithInfo(info interface{}) Sequence
}

type withFeatures interface {
	WithFeatures(ff FeatureTable) Sequence
}

type withBytes interface {
	WithBytes(p []byte) Sequence
}

// WithInfo creates a shallow copy of the given Sequence object and swaps the
// metadata with the given value. If the sequence implements the
// `WithInfo(info interface{}) Sequence` method, it will be called instead.
func WithInfo(seq Sequence, info interface{}) Sequence {
	switch v := seq.(type) {
	case withInfo:
		return v.WithInfo(info)
	default:
		return New(info, seq.Features(), seq.Bytes())
	}
}

// WithFeatures creates a shallow copy of the given Sequence object and swaps
// the feature table with the given features. If the sequence implements the
// `WithFeatures(ff FeatureTable) Sequence` method, it will be called instead.
func WithFeatures(seq Sequence, ff []Feature) Sequence {
	switch v := seq.(type) {
	case withFeatures:
		return v.WithFeatures(ff)
	default:
		return New(seq.Info(), ff, seq.Bytes())
	}
}

// WithBytes creates a shallow copy of the given Sequence object and swaps the
// byte representation with the given byte slice. If the sequence implements the
// `WithBytes(p []info) Sequence` method, it will be called instead.
func WithBytes(seq Sequence, p []byte) Sequence {
	switch v := seq.(type) {
	case withBytes:
		return v.WithBytes(p)
	default:
		return New(seq.Info(), seq.Features(), p)
	}
}

func insert(p []byte, pos int, q []byte) []byte {
	return append(p[:pos], append(q, p[pos:]...)...)
}

// Insert a sequence at the given position. For any feature whose location
// covers a region containing the point of insertion, the location will be
// split at the positions before and after the guest sequence.
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

// Embed a sequence at the given position. For any feature whose location
// covers a region containing the point of insertion, the location will be
// extended by the length of the guest Sequence.
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

// Delete a region of the sequence at the given position and length. Any
// features with a location containing the point of deletion will be
// shortened by the length of deletion. If the entirety of the feature is
// shortened as a result, the location will be described as a position in
// between the bases where the deletion occurred.
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
// to end. The target sequence region is copied. Any features with locations
// overlapping with the sliced region will be left in the sliced sequence.
func Slice(seq Sequence, start, end int) Sequence {
	p := make([]byte, end-start)
	copy(p, seq.Bytes()[start:end])
	var ff []Feature
	for _, f := range seq.Features() {
		loc := f.Location.Expand(0, -start).Expand(end-start, end-Len(seq))
		if loc.Len() != 0 || f.Location.Len() == 0 {
			f.Location = loc
			ff = append(ff, f)
		}
	}
	return WithBytes(WithFeatures(WithTopology(seq, Linear), ff), p)
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
// reversed order. The feature locations will be reversed accordingly.
func Reverse(seq Sequence) Sequence {
	var ff FeatureTable
	for _, f := range seq.Features() {
		ff = ff.Insert(Feature{f.Key, f.Location.Reverse(Len(seq)), f.Qualifiers, f.Order})
	}
	p := make([]byte, Len(seq))
	copy(p, seq.Bytes())
	flip.Bytes(p)
	return WithBytes(WithFeatures(seq, ff), p)
}

// Rotate returns a Sequence object whose coordinates are shifted by the given
// amount. Features which surpass the representational edges of the sequences
// are shifted and split as necessary.
func Rotate(seq Sequence, n int) Sequence {
	for Len(seq) > 0 && n < 0 {
		n += Len(seq)
	}
	var ff FeatureTable
	for _, f := range seq.Features() {
		loc := f.Location.Expand(0, n).Normalize(Len(seq))
		ff = ff.Insert(Feature{f.Key, loc, f.Qualifiers, f.Order})
	}
	p := seq.Bytes()
	p = append(p[n:], p[:n]...)
	return WithBytes(WithFeatures(seq, ff), p)
}
