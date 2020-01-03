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

// Len returns the length of a Sequence.
func Len(seq Sequence) int { return len(seq.Bytes()) }

// MutableSequence is a sequence which can be manipulated in-place.
type MutableSequence interface {
	Insert(pos int, seq Sequence) error
	Delete(pos, cnt int) error
	Replace(pos int, seq Sequence) error
	Sequence
}

type modifier interface {
	Apply(mut MutableSequence) error
}

type insertModifier struct {
	Pos int
	Seq Sequence
}

func (ins insertModifier) Apply(mut MutableSequence) error {
	return mut.Insert(ins.Pos, ins.Seq)
}

type deleteModifier struct {
	Pos int
	Cnt int
}

func (del deleteModifier) Apply(mut MutableSequence) error {
	return mut.Delete(del.Pos, del.Cnt)
}

type replaceModifier struct {
	Pos int
	Seq Sequence
}

func (rep replaceModifier) Apply(mut MutableSequence) error {
	return mut.Replace(rep.Pos, rep.Seq)
}

// BasicSequence is the most basic sequence type available. It is merely an
// alternate definition of a byte slice.
type BasicSequence []byte

// Bytes satisifies the gts.Sequence interface.
func (s BasicSequence) Bytes() []byte { return []byte(s) }

// Insert a sequence at the specified position.
func (s *BasicSequence) Insert(pos int, seq Sequence) error {
	if len(*s) <= pos {
		return fmt.Errorf(
			"unable to insert to sequence with length [%d] at position [%d]",
			len(*s), pos,
		)
	}
	if n := Len(seq); n > 0 {
		p := make([]byte, len(*s)+n)
		copy(p, (*s)[:pos])
		copy(p[pos:], seq.Bytes())
		copy(p[pos+n:], (*s)[pos:])
		*s = p
	}
	return nil
}

// Delete given number of bases from the specified position.
func (s *BasicSequence) Delete(pos, cnt int) error {
	if len(*s) <= pos {
		return fmt.Errorf(
			"unable to delete from sequence with length [%d] from position [%d]",
			len(*s), pos,
		)
	}
	if len(*s) <= pos+cnt {
		return fmt.Errorf(
			"unable to delete from sequence with length [%d] to position [%d]",
			len(*s), pos+cnt,
		)
	}
	if cnt > 0 {
		*s = append((*s)[:pos], (*s)[pos+cnt:]...)
	}
	return nil
}

// Replace the bases from the specified position with the given sequence.
func (s *BasicSequence) Replace(pos int, seq Sequence) error {
	if len(*s) <= pos {
		return fmt.Errorf(
			"unable to replace sequence with length [%d] from position [%d]",
			len(*s), pos,
		)
	}
	if len(*s) <= pos+Len(seq) {
		return fmt.Errorf(
			"unable to replace sequence with length [%d] to position [%d]",
			len(*s), pos+Len(seq),
		)
	}
	if n := Len(seq); n > 0 {
		copy((*s)[pos:], seq.Bytes())
	}
	return nil
}

// Seq is a convenience utility function which will take the given argument
// and attempt to convert the object into a BasicSequence.
func Seq(arg interface{}) BasicSequence {
	switch v := arg.(type) {
	case BasicSequence:
		return v
	case Sequence:
		return Seq(v.Bytes())
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

// Slice returns a slice of the Sequnece.
func Slice(seq Sequence, start, end int) BasicSequence {
	return BasicSequence(seq.Bytes()[start:end])
}

// Fragment returns a slice of Sequences containing all subsequences of the
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

// SequenceProxy represents a mutable sequence that will send modification
// requests to a SequenceServer.
type SequenceProxy struct {
	reqch chan bool
	seqch chan []byte
	modch chan modifier
	errch chan error
}

// Bytes satisfies the gts.Sequence interface.
func (sp SequenceProxy) Bytes() []byte {
	sp.reqch <- true
	return <-sp.seqch
}

// Insert a sequence at the specified position.
func (sp SequenceProxy) Insert(pos int, seq Sequence) error {
	sp.modch <- insertModifier{pos, seq}
	return <-sp.errch
}

// Delete given number of bases from the specified position.
func (sp SequenceProxy) Delete(pos, cnt int) error {
	sp.modch <- deleteModifier{pos, cnt}
	return <-sp.errch
}

// Replace the bases from the specified position with the given sequence.
func (sp SequenceProxy) Replace(pos int, seq Sequence) error {
	sp.modch <- replaceModifier{pos, seq}
	return <-sp.errch
}

// SequenceServer is a mutable sequence that can be modified by proxies.
type SequenceServer struct {
	mut   MutableSequence
	reqch chan bool
	seqch chan []byte
	modch chan modifier
	errch chan error
	spun  bool
}

// NewSequenceServer creates a new SequenceServer.
func NewSequenceServer(seq Sequence) SequenceServer {
	if mut, ok := seq.(MutableSequence); ok {
		ss := SequenceServer{
			mut,
			make(chan bool),
			make(chan []byte),
			make(chan modifier),
			make(chan error),
			false,
		}
		go ss.Spin()
		return ss
	}
	p := seq.Bytes()
	return NewSequenceServer((*BasicSequence)(&p))
}

// Spin prepares internal channels for recieving messages from the proxies
// associated to this SequenceServer.
func (ss *SequenceServer) Spin() {
	if !ss.spun {
		ss.spun = true
		for ss.spun {
			select {
			case mod := <-ss.modch:
				ss.errch <- mod.Apply(ss.mut)
			case flag := <-ss.reqch:
				if flag {
					ss.seqch <- ss.mut.Bytes()
				} else {
					ss.spun = false
				}
			}
		}
	}
}

// Close deactivates the SequenceServer.
func (ss *SequenceServer) Close() { ss.reqch <- false }

// Proxy creates a proxy to the SequenceServer.
func (ss SequenceServer) Proxy() SequenceProxy {
	return SequenceProxy{ss.reqch, ss.seqch, ss.modch, ss.errch}
}

// Bytes satisfies the gts.Sequence interface.
func (ss SequenceServer) Bytes() []byte {
	ss.reqch <- true
	return <-ss.seqch
}

// Insert a sequence at the specified position.
func (ss *SequenceServer) Insert(pos int, seq Sequence) error {
	return ss.mut.Insert(pos, seq)
}

// Delete given number of bases from the specified position.
func (ss *SequenceServer) Delete(pos, cnt int) error {
	return ss.mut.Delete(pos, cnt)
}

// Replace the bases from the specified position with the given sequence.
func (ss *SequenceServer) Replace(pos int, seq Sequence) error {
	return ss.mut.Replace(pos, seq)
}
