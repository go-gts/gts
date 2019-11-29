package gt1

import (
	"fmt"
	"strings"
)

type Sequence interface {
	Bytes() []byte
}

type Mutable interface {
	Sequence
	Insert(pos int, seq Sequence)
	Delete(pos, cnt int)
	Replace(pos int, seq Sequence)
}

type seqType []byte

func (seq seqType) Bytes() []byte { return []byte(seq) }

func Seq(arg interface{}) Sequence {
	switch v := arg.(type) {
	case Sequence:
		return v
	case []byte:
		return seqType(v)
	case string:
		return Seq([]byte(v))
	case []rune:
		return Seq(string(v))
	default:
		panic(fmt.Sprintf("cannot interpret object of type `%T` as a sequence", v))
	}
}

func Slice(seq Sequence, start, end int) Sequence {
	return Seq(seq.Bytes()[start:end])
}

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
