package gts

import (
	"bytes"
	"fmt"

	"github.com/go-flip/flip"
)

// Sequence represents an immutable sequence.
type Sequence interface {
	Bytes() []byte
}

// Operation represents a sequence operation.
type Operation func(seq Sequence) Sequence

// Apply the given operations to a sequence.
func Apply(seq Sequence, ops ...Operation) Sequence {
	for _, op := range ops {
		seq = op(seq)
	}
	return seq
}

// Len returns the length of the sequence.
func Len(seq Sequence) int {
	switch v := seq.(type) {
	case interface{ Len() int }:
		return v.Len()
	default:
		return len(v.Bytes())
	}
}

// Equals returns true if the sequences are identical.
func Equals(lhs, rhs Sequence) bool {
	return bytes.Equal(lhs.Bytes(), rhs.Bytes())
}

// Copy the given sequence.
func Copy(seq Sequence) Sequence {
	p := seq.Bytes()
	q := make(Seq, len(p))
	copy(q, p)
	return q
}

// Sliceable represents a sequence that can be sliced.
type Sliceable interface {
	Slice(i, j int) Sequence
}

// Slice returns a slice of the sequence.
func Slice(seq Sequence, i, j int) Sequence {
	switch v := seq.(type) {
	case Sliceable:
		return v.Slice(i, j)
	default:
		return Seq(seq.Bytes()[i:j])
	}
}

// Slicer returns a sequence slice operation.
func Slicer(i, j int) Operation {
	return func(seq Sequence) Sequence {
		return Slice(seq, i, j)
	}
}

// Insertable represents a sequence that can be inserted.
type Insertable interface {
	Insert(pos int, arg Sequence) Sequence
}

// Insert the given sequence into this sequence at the given position.
func Insert(seq Sequence, pos int, arg Sequence) Sequence {
	switch v := seq.(type) {
	case Insertable:
		return v.Insert(pos, arg)
	default:
		p := Copy(seq).Bytes()
		return Seq(append(p[:pos], append(arg.Bytes(), p[pos:]...)...))
	}
}

// Inserter returns a sequence insert operation.
func Inserter(pos int, arg Sequence) Operation {
	return func(seq Sequence) Sequence {
		return Insert(seq, pos, arg)
	}
}

// Replaceable represents a sequence that can be replaced.
type Replaceable interface {
	Replace(pos int, arg Sequence) Sequence
}

// Replace the bytes at the given position with the given sequence.
func Replace(seq Sequence, pos int, arg Sequence) Sequence {
	switch v := seq.(type) {
	case Replaceable:
		return v.Replace(pos, arg)
	default:
		p := Copy(seq).Bytes()
		copy(p[pos:], arg.Bytes())
		return Seq(p)
	}
}

// Replacer returns a sequence replace operation.
func Replacer(pos int, arg Sequence) Operation {
	return func(seq Sequence) Sequence {
		return Replace(seq, pos, arg)
	}
}

// Deletable represents a sequence that can be deleted.
type Deletable interface {
	Delete(pos, arg int) Sequence
}

// Delete the given number of bytes at the given position.
func Delete(seq Sequence, pos, arg int) Sequence {
	switch v := seq.(type) {
	case Deletable:
		return v.Delete(pos, arg)
	default:
		p := Copy(seq).Bytes()
		return Seq(append(p[:pos], p[pos+arg:]...))
	}
}

// Deleter returns a sequence replace operation.
func Deleter(pos, arg int) Operation {
	return func(seq Sequence) Sequence {
		return Delete(seq, pos, arg)
	}
}

// Seq represents a bare byte slice sequence.
type Seq []byte

// AsSequence returns a Sequence object represented by a byte slice.
func AsSequence(arg interface{}) Sequence {
	switch v := arg.(type) {
	case Sequence:
		return v
	case []byte:
		return Seq(v)
	case []rune:
		return Seq(string(v))
	case string:
		return Seq(v)
	case byte:
		return Seq{v}
	case rune:
		return Seq(string([]rune{v}))
	default:
		panic(fmt.Sprintf("invalid object in New: %T", v))
	}
}

// Bytes returns the byte representation of the sequence.
func (seq Seq) Bytes() []byte {
	return seq
}

// Reverse returns a sequence in the reversed order.
func Reverse(arg Sequence) Sequence {
	p := Copy(arg).Bytes()
	flip.Bytes(p)
	return Seq(p)
}

// Concat concatenates the given sequences.
func Concat(seqs ...Sequence) Sequence {
	n := 0
	for _, seq := range seqs {
		n += Len(seq)
	}

	p := make([]byte, n)
	i := 0
	for _, seq := range seqs {
		i += copy(p[i:], seq.Bytes())
	}

	return Seq(p)
}
