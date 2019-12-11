package gts

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

// BasicSequence is the most basic sequence type available. It is merely an
// alternate definition of a byte slice.
type BasicSequence []byte

// Bytes satisifies the gts.Sequence interface.
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

// Mutable is a sequence which can be manipulated in-place.
type Mutable interface {
	Sequence
	Insert(pos int, seq Sequence)
	Delete(pos, cnt int)
	Replace(pos int, seq Sequence)
}

type insArg struct {
	Pos int
	Seq Sequence
}

type delArg struct {
	Pos int
	Cnt int
}

type repArg struct {
	Pos int
	Seq Sequence
}

// SequenceServer represents a sequence that can be modified by proxies.
type SequenceServer struct {
	data  []byte
	insch chan insArg
	delch chan delArg
	repch chan repArg
	locch chan Location
	seqch chan Sequence

	closech  chan interface{}
	spinning bool
}

// NewSequenceServer creates a new SequenceServer.
func NewSequenceServer(seq Sequence) *SequenceServer {
	ss := &SequenceServer{
		seq.Bytes(),
		make(chan insArg),
		make(chan delArg),
		make(chan repArg),
		make(chan Location),
		make(chan Sequence),
		make(chan interface{}),
		false,
	}
	go ss.Spin()
	return ss
}

// Spin the server.
func (ss *SequenceServer) Spin() {
	ss.spinning = true
	for ss.spinning {
		select {
		case msg := <-ss.insch:
			ss.Insert(msg.Pos, msg.Seq)
		case msg := <-ss.delch:
			ss.Delete(msg.Pos, msg.Cnt)
		case msg := <-ss.repch:
			ss.Insert(msg.Pos, msg.Seq)
		case loc := <-ss.locch:
			ss.seqch <- loc.Locate(ss)
		case <-ss.closech:
			ss.spinning = false
		}
	}
}

// Close the server.
func (ss *SequenceServer) Close() {
	ss.closech <- nil
}

// Bytes satisfies the gts.Sequence interface.
func (ss SequenceServer) Bytes() []byte { return ss.data }

// Insert a sequence at the specified position.
func (ss *SequenceServer) Insert(pos int, seq Sequence) {
	if n := len(seq.Bytes()); n > 0 {
		p := make([]byte, len(ss.data)+n)
		copy(p, ss.data[:pos])
		copy(p[pos:], seq.Bytes())
		copy(p[pos+n:], ss.data[pos:])
	}
}

// Delete given number of bases from the specified position.
func (ss *SequenceServer) Delete(pos, cnt int) {
	if cnt > 0 {
		ss.data = append(ss.data[:pos], ss.data[pos+cnt:]...)
	}
}

// Replace the bases from the specified position with the given sequence.
func (ss *SequenceServer) Replace(pos int, seq Sequence) {
	if n := len(seq.Bytes()); n > 0 {
		copy(ss.data[pos:], seq.Bytes())
	}
}

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
