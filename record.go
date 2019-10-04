package gt1

import (
	"time"
)

type Organism struct {
	Species string
	Name    string
	Taxon   []string
}

type Reference struct {
	Number  int
	Start   int
	End     int
	Authors string
	Group   string
	Title   string
	Journal string
	Xref    map[string]string
	Comment string
}

type Feature struct {
	Key        string
	Location   Location
	Properties PairList
}

func (f Feature) Insert() {
}

type Record struct {
	LocusName  string
	Accessions []string
	Topology   string
	Version    string
	Molecule   string
	Class      string
	Division   string
	Dates      []time.Time
	DBLink     PairList

	Definition string
	Keywords   []string

	Source     Organism
	References []Reference
	Comment    string
	Features   []Feature

	s []byte

	insch chan insArg
	delch chan delArg
	repch chan repArg
	locch chan Location
	seqch chan Sequence
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

func NewRecord() *Record {
	r := &Record{}
	r.Start()
	return r
}

func (record *Record) Start() {
	go func() {
		for {
			select {
			case msg := <-record.insch:
				record.Insert(msg.Pos, msg.Seq)
			case msg := <-record.delch:
				record.Delete(msg.Pos, msg.Cnt)
			case msg := <-record.repch:
				record.Insert(msg.Pos, msg.Seq)
			case loc := <-record.locch:
				record.seqch <- loc.Locate(record)
			}
		}
	}()
}

func (record Record) Bytes() []byte {
	return record.s
}

func (record Record) String() string {
	return string(record.s)
}

func (record Record) Length() int {
	return len(record.s)
}

func (record Record) Slice(start, end int) Sequence {
	for start < 0 {
		start += len(record.s)
	}
	for end < 0 {
		end += len(record.s)
	}
	return Seq(record.s[start:end])
}

func (record Record) Subseq(loc Location) Sequence {
	return loc.Locate(record)
}

func (record *Record) Insert(pos int, seq Sequence) {
	record.s = insertBytes(record.s, pos, seq.Bytes())
	for _, feature := range record.Features {
		feature.Location.Shift(pos, seq.Length())
	}
}

func (record *Record) Delete(pos, cnt int) {
	record.s = deleteBytes(record.s, pos, cnt)
	for _, feature := range record.Features {
		feature.Location.Shift(pos+cnt, -cnt)
	}
}

func (record *Record) Replace(pos int, seq Sequence) {
	record.s = replaceBytes(record.s, pos, seq.Bytes())
}

func insertBytes(s []byte, pos int, vs []byte) []byte {
	r := make([]byte, len(s)+len(vs))
	copy(r[:pos], s[:pos])
	copy(r[pos:], vs)
	copy(r[pos+len(vs):], s[pos:])
	return r
}

func deleteBytes(s []byte, pos, cnt int) []byte {
	r := make([]byte, len(s)-cnt)
	copy(r[:pos], s[:pos])
	copy(r[pos:], s[pos+cnt:])
	return r
}

func replaceBytes(s []byte, pos int, vs []byte) []byte {
	r := make([]byte, len(s))
	copy(r[:pos], s[:pos])
	copy(r[pos:], vs)
	copy(r[pos+len(vs):], s[pos+len(vs):])
	return r
}
