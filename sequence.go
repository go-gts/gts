package gts

import (
	"bytes"
	"index/suffixarray"
	"reflect"
	"sort"

	"github.com/go-flip/flip"
)

// Shiftable represents a shiftable metadata.
type Shiftable interface {
	Shift(i, n int) interface{}
}

// Expandable represents a expandable metadata.
type Expandable interface {
	Expand(i, n int) interface{}
}

// Sliceable represents a sliceable metadata.
type Sliceable interface {
	Slice(start, end int) interface{}
}

func tryShift(info interface{}, i, n int) interface{} {
	if v, ok := info.(Shiftable); ok {
		return v.Shift(i, n)
	}
	return info
}

func tryExpand(info interface{}, i, n int) interface{} {
	if v, ok := info.(Expandable); ok {
		return v.Expand(i, n)
	}
	return info
}

func trySlice(info interface{}, start, end int) interface{} {
	if v, ok := info.(Sliceable); ok {
		return v.Slice(start, end)
	}
	return info
}

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

type hasWithInfo interface {
	WithInfo(info interface{}) Sequence
}

type hasWithFeatures interface {
	WithFeatures(ff FeatureTable) Sequence
}

type hasWithBytes interface {
	WithBytes(p []byte) Sequence
}

// WithInfo creates a shallow copy of the given Sequence object and swaps the
// metadata with the given value. If the sequence implements the
// `WithInfo(info interface{}) Sequence` method, it will be called instead.
func WithInfo(seq Sequence, info interface{}) Sequence {
	switch v := seq.(type) {
	case hasWithInfo:
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
	case hasWithFeatures:
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
	case hasWithBytes:
		return v.WithBytes(p)
	default:
		return New(seq.Info(), seq.Features(), p)
	}
}

func insert(p []byte, pos int, q []byte) []byte {
	return append(p[:pos], append(q, p[pos:]...)...)
}

// Insert a sequence at the given index. For any feature whose location covers
// a region containing the point of insertion, the location will be split at
// the positions before and after the guest sequence.
func Insert(host Sequence, index int, guest Sequence) Sequence {
	info := host.Info()
	info = tryShift(info, index, Len(guest))
	host = WithInfo(host, info)

	var ff FeatureTable
	for _, f := range host.Features() {
		f.Location = f.Location.Shift(index, Len(guest))
		ff = ff.Insert(f)
	}
	for _, f := range guest.Features() {
		f.Location = f.Location.Expand(0, index)
		ff = ff.Insert(f)
	}
	host = WithFeatures(host, ff)

	p := insert(host.Bytes(), index, guest.Bytes())
	host = WithBytes(host, p)

	return host
}

// Embed a sequence at the given index. For any feature whose location covers
// a region containing the point of insertion, the location will be extended
// by the length of the guest Sequence.
func Embed(host Sequence, index int, guest Sequence) Sequence {
	info := host.Info()
	info = tryExpand(info, index, Len(guest))
	host = WithInfo(host, info)

	var ff FeatureTable
	for _, f := range host.Features() {
		f.Location = f.Location.Expand(index, Len(guest))
		ff = ff.Insert(f)
	}
	for _, f := range guest.Features() {
		f.Location = f.Location.Expand(0, index)
		ff = ff.Insert(f)
	}
	host = WithFeatures(host, ff)

	p := insert(host.Bytes(), index, guest.Bytes())
	host = WithBytes(host, p)

	return host
}

// Delete a region of the sequence at the given offset and length. Any
// features with a location containing the point of deletion will be
// shortened by the length of deletion. If the entirety of the feature is
// shortened as a result, the location will be described as a offset in
// between the bases where the deletion occurred.
func Delete(seq Sequence, offset, length int) Sequence {
	info := seq.Info()
	info = tryExpand(info, offset, -length)
	seq = WithInfo(seq, info)

	ff := make([]Feature, len(seq.Features()))
	for i, f := range seq.Features() {
		ff[i].Key = f.Key
		ff[i].Location = f.Location.Expand(offset, -length)
		ff[i].Qualifiers = f.Qualifiers
	}
	seq = WithFeatures(seq, ff)

	q := seq.Bytes()
	p := make([]byte, len(q)-length)
	copy(p[:offset], q[:offset])
	copy(p[offset:], q[offset+length:])
	seq = WithBytes(seq, p)

	return seq
}

// Erase a region of the sequence at the given offset and length. Any
// features with a location containing the point of deletion will be
// shortened by the length of deletion. If the entirety of the feature is
// shortened as a result, the location will be removed from the sequence.
func Erase(seq Sequence, offset, length int) Sequence {
	f := Or(Key("source"), Not(Within(offset, offset+length)))
	ff := seq.Features().Filter(f)
	seq = WithFeatures(seq, ff)
	return Delete(seq, offset, length)
}

// Slice returns a subsequence of the given sequence starting at start and up
// to end. The target sequence region is copied. Any features with locations
// overlapping with the sliced region will be left in the sliced sequence.
func Slice(seq Sequence, start, end int) Sequence {
	seqlen := Len(seq)
	if start < 0 {
		start += seqlen
	}

	if end < 0 {
		end += seqlen
	}

	if end < start {
		length := seqlen - start + end
		seq = Rotate(seq, -start)
		return Slice(seq, 0, length)
	}

	info := seq.Info()
	info = trySlice(info, start, end)

	ff := seq.Features().Filter(Overlap(start, end))

	for i, f := range ff {
		loc := f.Location.Expand(end, end-seqlen).Expand(0, -start)
		if f.Key == "source" {
			loc = asComplete(loc)
		}
		ff[i].Location = loc
	}

	p := make([]byte, end-start)
	copy(p, seq.Bytes()[start:end])

	seq = WithInfo(seq, info)
	seq = WithFeatures(seq, ff)
	seq = WithBytes(seq, p)
	seq = WithTopology(seq, Linear)

	return seq
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

		head = WithFeatures(head, ff)
		head = WithBytes(head, p)

		return head
	}
}

// Reverse returns a Sequence object with the byte representation in the
// reversed order. The feature locations will be reversed accordingly.
func Reverse(seq Sequence) Sequence {
	var ff FeatureTable
	for _, f := range seq.Features() {
		ff = ff.Insert(Feature{
			f.Key,
			f.Location.Reverse(Len(seq)),
			f.Qualifiers,
			f.Order,
		})
	}
	seq = WithFeatures(seq, ff)

	p := make([]byte, Len(seq))
	copy(p, seq.Bytes())
	flip.Bytes(p)
	seq = WithBytes(seq, p)

	return seq
}

// Rotate returns a Sequence object whose coordinates are shifted by the given
// amount. Features which surpass the representational edges of the sequences
// are shifted and split as necessary.
func Rotate(seq Sequence, n int) Sequence {
	for Len(seq) > 0 && n < 0 {
		n += Len(seq)
	}
	n %= Len(seq)

	var ff FeatureTable
	for _, f := range seq.Features() {
		loc := f.Location.Expand(0, n).Normalize(Len(seq))
		ff = ff.Insert(Feature{f.Key, loc, f.Qualifiers, f.Order})
	}

	m := Len(seq) - n
	p := seq.Bytes()
	p = append(p[m:], p[:m]...)

	seq = WithFeatures(seq, ff)
	seq = WithBytes(seq, p)

	return seq
}

func bytesIndexAll(s, sep []byte) []int {
	index := suffixarray.New(s)
	return index.Lookup(sep, -1)
}

// Search for a subsequence within a sequence.
func Search(seq Sequence, query Sequence) []Segment {
	if Len(seq) == 0 || Len(query) == 0 {
		return nil
	}

	s := bytes.ToLower(seq.Bytes())
	sep := bytes.ToLower(query.Bytes())

	indices := bytesIndexAll(s, sep)
	segments := make([]Segment, len(indices))
	for i, index := range indices {
		segments[i] = Segment{index, index + len(sep)}
	}
	sort.Sort(BySegment(segments))
	return segments
}
