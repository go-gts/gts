package gd

import "fmt"

type Sequence interface {
	View(start, end int) Sequence
	Bytes() []byte
	String() string
	Length() int
	Insert(pos int, seq Sequence)
	Delete(pos, count int)
	Replace(pos int, seq Sequence)
}

type SeqLike interface{}

func Seq(s SeqLike) Sequence {
	switch v := s.(type) {
	case []byte:
		seq := &sequence{
			bytes: []byte(v),
			strch: make(chan []byte),
			reqch: make(chan seqRange),
			opch:  make(chan seqOp),
		}
		go seq.Start()
		return seq
	case string:
		return Seq([]byte(v))
	case []rune:
		return Seq([]byte(string(v)))
	default:
		panic(fmt.Errorf("cannot make a sequence from `%T`", v))
	}
}

func insertFunc(s []byte, pos int, vs []byte) []byte {
	r := make([]byte, len(s)+len(vs))
	copy(r[:pos], s[:pos])
	copy(r[pos:], vs)
	copy(r[pos+len(vs):], s[pos:])
	return r
}

func deleteFunc(s []byte, pos, count int) []byte {
	r := make([]byte, len(s)-count)
	copy(r[:pos], s[:pos])
	copy(r[pos:], s[pos+count:])
	return r
}

func replaceFunc(s []byte, pos int, vs []byte) []byte {
	r := make([]byte, len(s))
	copy(r, s)
	copy(r[pos:], vs)
	return r
}

type seqRange struct {
	Start int
	End   int
}

type seqOp interface {
	Apply([]byte) []byte
}

type insertOp struct {
	Position int
	Value    []byte
}

func newInsertOp(pos int, value []byte) seqOp {
	return insertOp{Position: pos, Value: value}
}

func (op insertOp) Apply(s []byte) []byte {
	return insertFunc(s, op.Position, op.Value)
}

type deleteOp struct {
	Position int
	Count    int
}

func newDeleteOp(pos, count int) seqOp {
	return deleteOp{Position: pos, Count: count}
}

func (op deleteOp) Apply(s []byte) []byte {
	return deleteFunc(s, op.Position, op.Count)
}

type replaceOp struct {
	Position int
	Value    []byte
}

func newReplaceOp(pos int, value []byte) seqOp {
	return replaceOp{Position: pos, Value: value}
}

func (op replaceOp) Apply(s []byte) []byte {
	return replaceFunc(s, op.Position, op.Value)
}

type sequence struct {
	bytes []byte
	strch chan []byte
	reqch chan seqRange
	opch  chan seqOp
}

func (s *sequence) Start() {
	for {
		select {
		case req := <-s.reqch:
			s.strch <- s.bytes[req.Start:req.End]
		case op := <-s.opch:
			s.bytes = op.Apply(s.bytes)
		}
	}
}

func (s sequence) View(start, end int) Sequence {
	if start < 0 {
		start += s.Length()
	}
	if end < 0 {
		end += s.Length()
	}
	if end < start {
		panic(fmt.Errorf("runtime error: View start %d is smaller than end %d", start, end))
	}
	if end > s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", end, s.Length()))
	}
	return &seqview{
		start: start,
		end:   end,
		strch: s.strch,
		reqch: s.reqch,
		opch:  s.opch,
	}
}

func (s sequence) Bytes() []byte {
	return s.bytes
}

func (s sequence) String() string {
	return string(s.bytes)
}

func (s sequence) Length() int {
	return len(s.bytes)
}

func (s *sequence) Insert(pos int, seq Sequence) {
	if pos < 0 {
		pos += s.Length()
	}
	if pos >= s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", pos, s.Length()))
	}
	s.bytes = insertFunc(s.bytes, pos, seq.Bytes())
}

func (s *sequence) Delete(pos, count int) {
	if pos < 0 {
		pos += s.Length()
	}
	if pos >= s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", pos, s.Length()))
	}
	s.bytes = deleteFunc(s.bytes, pos, count)
}

func (s *sequence) Replace(pos int, seq Sequence) {
	if pos < 0 {
		pos += s.Length()
	}
	if pos >= s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", pos, s.Length()))
	}
	s.bytes = replaceFunc(s.bytes, pos, seq.Bytes())
}

type seqview struct {
	start int
	end   int
	strch chan []byte
	reqch chan seqRange
	opch  chan seqOp
}

func (s seqview) View(start, end int) Sequence {
	if start < 0 {
		start += s.Length()
	}
	if end < 0 {
		end += s.Length()
	}
	if end < start {
		panic(fmt.Errorf("runtime error: View start %d is smaller than end %d", start, end))
	}
	if end > s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", end, s.Length()))
	}
	return &seqview{
		start: s.start + start,
		end:   s.start + end,
		strch: s.strch,
		reqch: s.reqch,
		opch:  s.opch,
	}
}

func (s seqview) Bytes() []byte {
	s.reqch <- seqRange{Start: s.start, End: s.end}
	return <-s.strch
}

func (s seqview) String() string {
	return string(s.Bytes())
}

func (s seqview) Length() int {
	return s.end - s.start
}

func (s *seqview) Insert(pos int, seq Sequence) {
	if pos < 0 {
		pos += s.Length()
	}
	if pos >= s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", pos, s.Length()))
	}
	s.opch <- newInsertOp(s.start+pos, seq.Bytes())
	s.end += seq.Length()
}

func (s *seqview) Delete(pos, count int) {
	if pos < 0 {
		pos += s.Length()
	}
	if pos >= s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", pos, s.Length()))
	}
	s.opch <- newDeleteOp(s.start+pos, count)
	s.end -= count
}

func (s seqview) Replace(pos int, seq Sequence) {
	if pos < 0 {
		pos += s.Length()
	}
	if pos >= s.Length() {
		panic(fmt.Errorf("runtime error: index out of range [%d] with length %d", pos, s.Length()))
	}
	s.opch <- newReplaceOp(s.start+pos, seq.Bytes())
}
