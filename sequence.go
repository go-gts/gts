package gt1

import (
	"bytes"
	"fmt"
	"strings"
)

// Sequence represents a sequence type. All Sequence types must be able to
// generate a byte slice representation of itself.
type Sequence interface {
	Bytes() []byte
}

// Mutable is a sequence which can be manipulated in-place.
type Mutable interface {
	Sequence
	Insert(pos int, seq Sequence)
	Delete(pos, cnt int)
	Replace(pos int, seq Sequence)
}

// BasicSequence is the most basic sequence type available. It is merely an
// alternate definition of a byte slice.
type BasicSequence []byte

// Bytes satisifies the gt1.Sequence interface.
func (seq BasicSequence) Bytes() []byte { return []byte(seq) }

// Seq is a convenience utility function which will take the given argument
// and attempt to convert the object into a Sequence.
func Seq(arg interface{}) Sequence {
	switch v := arg.(type) {
	case Sequence:
		return v
	case []byte:
		return BasicSequence(v)
	case string:
		return Seq([]byte(v))
	case []rune:
		return Seq(string(v))
	case fmt.Stringer:
		return Seq(v.String())
	default:
		panic(fmt.Sprintf("cannot interpret object of type `%T` as a sequence", v))
	}
}

// Equal tests if the given Sequences have the smae byte slice representations.
func Equal(a, b Sequence) bool { return bytes.Equal(a.Bytes(), b.Bytes()) }

// Slice returns a slice of the Sequnece.
func Slice(seq Sequence, start, end int) Sequence {
	return Seq(seq.Bytes()[start:end])
}

// Fragment will return a slice of Sequences containing all subsequences of the
// given Sequence from position 0 separated by `slide` bytes and with length of
// `window`.
func Fragment(seq Sequence, window, slide int) []Sequence {
	p := seq.Bytes()
	ret := make([]Sequence, 0)
	for i := 0; i < len(p); i += slide {
		j := i + window
		if j > len(p) {
			j = len(p)
		}
		ret = append(ret, Seq(p[i:j]))
	}
	return ret
}

// Composition computes the occurence of each byte within the Sequence.
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

// Skew calculates the skewness of the Sequence. Where `n` is the number of
// bytes in the given Sequence also appearing in the `nSet` string and `p`
// is the number of bytes in the given Sequence also appearing in the `pSet`
// string, the skewness of the Sequence is calculated as: (p - n) / (p + n).
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
