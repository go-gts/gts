package gt1

import (
	"fmt"
	"strings"
)

type Sequence interface {
	// Bytes returns the raw representation of the sequence.
	Bytes() []byte

	// String returns the string representation of the sequence.
	String() string

	// Len returns the length of the sequence.
	Len() int

	// Slice returns the slice of the sequence.
	Slice(start, end int) Sequence

	// Subseq returns the subsequence to the given location.
	Subseq(loc Location) Sequence
}

type BytesLike interface{}

func AsBytes(s BytesLike) []byte {
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
	return seqType(AsBytes(s))
}

type seqType []byte

func (s seqType) Bytes() []byte {
	return []byte(s)
}

func (s seqType) String() string {
	return string(s)
}

func (s seqType) Len() int {
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
		l += seq.Len()
	}

	r := make([]byte, l)
	i := 0
	for _, seq := range seqs {
		copy(r[i:], seq.Bytes())
		i += seq.Len()
	}

	return Seq(r)
}

func Fragment(seq Sequence, window, slide int) []Sequence {
	ret := make([]Sequence, 0)
	for i := 0; i < seq.Len(); i += slide {
		j := i + window
		if j > seq.Len() {
			j = seq.Len()
		}
		fragment := seq.Slice(i, j)
		ret = append(ret, fragment)
	}
	return ret
}

func Composition(seq Sequence) map[byte]int {
	comp := make(map[byte]int)
	for _, b := range seq.Bytes() {
		if _, ok := comp[b]; !ok {
			comp[b] = 0
		}
		comp[b]++
	}
	return comp
}

func Skew(seq Sequence, nSet, pSet string) float64 {
	comp := Composition(seq)
	nCnt, pCnt := 0., 0.
	for b, n := range comp {
		v := float64(n)
		if strings.ContainsRune(nSet, rune(b)) {
			nCnt += v
		}
		if strings.ContainsRune(pSet, rune(b)) {
			pCnt += v
		}
	}
	return (pCnt - nCnt) / (pCnt + nCnt)
}
