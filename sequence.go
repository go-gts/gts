package gts

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type mutableByteSlice []byte

func (p *mutableByteSlice) Insert(pos int, arg Sequence) error {
	if len(*p) <= pos {
		return fmt.Errorf(
			"unable to insert at position [%d] for sequence of length [%d]",
			pos, len(*p),
		)
	}
	if n := Len(arg); n > 0 {
		q := make([]byte, len(*p)+n)
		copy(q, (*p)[:pos])
		copy(q[pos:], arg.Data())
		copy(q[pos+n:], (*p)[pos:])
		*p = q
	}
	return nil
}

func (p *mutableByteSlice) Delete(pos, arg int) error {
	if len(*p) <= pos {
		return fmt.Errorf(
			"unable to delete from position [%d] for sequence of length [%d]",
			pos, len(*p),
		)
	}
	if len(*p) <= pos+arg {
		return fmt.Errorf(
			"unable to delete to position [%d] for sequence of length [%d]",
			len(*p), pos+arg,
		)
	}
	if arg > 0 {
		*p = append((*p)[:pos], (*p)[pos+arg:]...)
	}
	return nil
}

func (p *mutableByteSlice) Replace(pos int, arg Sequence) error {
	if len(*p) <= pos {
		return fmt.Errorf(
			"unable to replace from position [%d] for sequence of length [%d]",
			pos, len(*p),
		)
	}
	if len(*p) <= pos+Len(arg) {
		return fmt.Errorf(
			"unable to replace to position [%d] for sequence of length [%d]",
			pos+Len(arg), len(*p),
		)
	}
	if n := Len(arg); n > 0 {
		copy((*p)[pos:], arg.Data())
	}
	return nil
}

// Sequence represents a sequence type. All Sequence types must be able to
// generate a byte slice representation of itself.
type Sequence interface {
	Info() interface{}
	Data() []byte
}

// Len returns the length of a Sequence.
func Len(seq Sequence) int { return len(seq.Data()) }

// Equal tests if the given Sequences have equal byte slice representations.
func Equal(a, b Sequence) bool { return bytes.Equal(a.Data(), b.Data()) }

// Identical tests if the given Sequences have identical metadata values and
// byte slice representations.
func Identical(a, b Sequence) bool {
	return reflect.DeepEqual(a.Info(), b.Info()) && Equal(a, b)
}

// MutableSequence is a sequence which can be manipulated in-place.
type MutableSequence interface {
	Insert(pos int, arg Sequence) error
	Delete(pos, arg int) error
	Replace(pos int, arg Sequence) error
	Sequence
}

// GTS represents the most basic GTS sequence object.
type GTS struct {
	info interface{}
	data []byte
}

// New creates a new GTS object.
func New(info interface{}, data []byte) *GTS {
	return &GTS{info, data}
}

// Seq converts an arbitrary value to a GTS object.
func Seq(v interface{}) *GTS {
	switch v := v.(type) {
	case *GTS:
		return v
	case Sequence:
		return New(v.Info(), v.Data())
	case []byte:
		return New(nil, v)
	case string:
		return Seq([]byte(v))
	default:
		panic(fmt.Sprintf("cannot interpret object of type `%T` as a sequence", v))
	}
}

// Info returns the metadata of the sequence.
func (seq GTS) Info() interface{} { return seq.info }

// Data returns the byte representation of the sequence.
func (seq GTS) Data() []byte { return []byte(seq.data) }

// Insert a sequence at the specified position.
func (seq *GTS) Insert(pos int, arg Sequence) error {
	return (*mutableByteSlice)(&(seq.data)).Insert(pos, arg)
}

// Delete given number of bases from the specified position.
func (seq *GTS) Delete(pos, arg int) error {
	return (*mutableByteSlice)(&(seq.data)).Delete(pos, arg)
}

// Replace the bases from the specified position with the given sequence.
func (seq *GTS) Replace(pos int, arg Sequence) error {
	return (*mutableByteSlice)(&(seq.data)).Replace(pos, arg)
}

// Slice returns a slice of the Sequnece.
func Slice(seq Sequence, start, end int) Sequence {
	return New(seq.Info(), seq.Data()[start:end])
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

type seqRequest int

const (
	closeRequest seqRequest = iota
	infoRequest
	dataRequest
)

// SequenceProxy represents a mutable sequence that will send modification
// requests to a SequenceServer.
type SequenceProxy struct {
	reqch  chan seqRequest
	infoch chan interface{}
	datach chan []byte
	modch  chan modifier
	errch  chan error
}

// Info satisfies the gts.Sequence interface.
func (sp SequenceProxy) Info() interface{} {
	sp.reqch <- infoRequest
	return <-sp.infoch
}

// Data satisfies the gts.Sequence interface.
func (sp SequenceProxy) Data() []byte {
	sp.reqch <- dataRequest
	return <-sp.datach
}

// Insert a sequence at the specified position.
func (sp SequenceProxy) Insert(pos int, arg Sequence) error {
	sp.modch <- insertModifier{pos, arg}
	return <-sp.errch
}

// Delete given number of bases from the specified position.
func (sp SequenceProxy) Delete(pos, arg int) error {
	sp.modch <- deleteModifier{pos, arg}
	return <-sp.errch
}

// Replace the bases from the specified position with the given sequence.
func (sp SequenceProxy) Replace(pos int, arg Sequence) error {
	sp.modch <- replaceModifier{pos, arg}
	return <-sp.errch
}

// SequenceServer is a mutable sequence that can be modified by proxies.
type SequenceServer struct {
	mut    MutableSequence
	reqch  chan seqRequest
	infoch chan interface{}
	datach chan []byte
	modch  chan modifier
	errch  chan error
	spun   bool
}

// NewSequenceServer creates a new SequenceServer.
func NewSequenceServer(seq Sequence) SequenceServer {
	if mut, ok := seq.(MutableSequence); ok {
		ss := SequenceServer{
			mut:    mut,
			reqch:  make(chan seqRequest),
			infoch: make(chan interface{}),
			datach: make(chan []byte),
			modch:  make(chan modifier),
			errch:  make(chan error),
			spun:   false,
		}
		go ss.Spin()
		return ss
	}
	return NewSequenceServer(New(seq.Info(), seq.Data()))
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
				switch flag {
				case closeRequest:
					ss.spun = false
				case infoRequest:
					ss.infoch <- ss.mut.Info()
				case dataRequest:
					ss.datach <- ss.mut.Data()
				}
			}
		}
	}
}

// Close deactivates the SequenceServer.
func (ss *SequenceServer) Close() { ss.reqch <- closeRequest }

// Proxy creates a proxy to the SequenceServer.
func (ss SequenceServer) Proxy() SequenceProxy {
	return SequenceProxy{ss.reqch, ss.infoch, ss.datach, ss.modch, ss.errch}
}

// Info satisfies the gts.Sequence interface.
func (ss SequenceServer) Info() interface{} {
	ss.reqch <- infoRequest
	return <-ss.infoch
}

// Data satisfies the gts.Sequence interface.
func (ss SequenceServer) Data() []byte {
	ss.reqch <- dataRequest
	return <-ss.datach
}

// Insert a sequence at the specified position.
func (ss *SequenceServer) Insert(pos int, arg Sequence) error {
	return ss.mut.Insert(pos, arg)
}

// Delete given number of bases from the specified position.
func (ss *SequenceServer) Delete(pos, arg int) error {
	return ss.mut.Delete(pos, arg)
}

// Replace the bases from the specified position with the given sequence.
func (ss *SequenceServer) Replace(pos int, arg Sequence) error {
	return ss.mut.Replace(pos, arg)
}

// Fragment returns a slice of Sequences containing all subsequences of the
// given Sequence from position 0 separated by `slide` bytes and with length of
// `window`.
func Fragment(seq Sequence, window, slide int) []Sequence {
	p := seq.Data()
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
	for _, b := range seq.Data() {
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
