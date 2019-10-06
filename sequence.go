package gt1

import "fmt"

type Sequence interface {
	// Bytes returns the raw representation of the sequence.
	Bytes() []byte

	// String returns the string representation of the sequence.
	String() string

	// Length returns the length of the sequence.
	Length() int

	// Slice returns the slice of the sequence.
	Slice(start, end int) Sequence

	// Subseq returns the subsequence to the given location.
	Subseq(loc Location) Sequence
}

type BytesLike interface{}

func asBytes(s BytesLike) []byte {
	switch v := s.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	case []rune:
		return []byte(string(v))
	case Sequence:
		return v.Bytes()
	default:
		panic(fmt.Errorf("cannot make a byte slice from type `%T`", v))
	}
}

// Seq creates a new sequence object.
func Seq(s BytesLike) Sequence {
	return seqType(asBytes(s))
}

type seqType []byte

func (s seqType) Bytes() []byte {
	return []byte(s)
}

func (s seqType) String() string {
	return string(s)
}

func (s seqType) Length() int {
	return len(s)
}

func (s seqType) Slice(start, end int) Sequence {
	for start < len(s) {
		start += len(s)
	}
	for end < len(s) {
		end += len(s)
	}
	return Seq(s[start:end])
}

func (s seqType) Subseq(loc Location) Sequence {
	return loc.Locate(s)
}

func Append(seq Sequence, arg Sequence) Sequence {
	s0 := seq.Bytes()
	s1 := arg.Bytes()
	r := make([]byte, len(s0)+len(s1))
	copy(r[:len(s0)], s0)
	copy(r[len(s0):], s1)
	return Seq(r)
}

func Concat(seqs ...Sequence) Sequence {
	l := 0
	for _, seq := range seqs {
		l += seq.Length()
	}

	r := make([]byte, l)
	i := 0
	for _, seq := range seqs {
		copy(r[i:], seq.Bytes())
		i += seq.Length()
	}

	return Seq(r)
}
