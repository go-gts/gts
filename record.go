package gt1

import (
	"io"
	"sort"
	"time"

	"github.com/ktnyt/pars"
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

type Metadata struct {
	LocusName  string
	Accessions []string
	Topology   string
	Version    string
	Molecule   string
	Class      string
	Division   string
	Dates      []time.Time
	DBLink     *PairList
	Definition string
	Keywords   []string
	Source     Organism
	References []Reference
	Comment    string
}

type Record interface {
	Fields() *Metadata
	Features() []Feature
	AddFeature(feature Feature)
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

type recordType struct {
	fields   *Metadata
	features []Feature

	origin []byte
	insch  chan insArg
	delch  chan delArg
	repch  chan repArg
	locch  chan Location
	seqch  chan Sequence
}

func NewRecord(fields *Metadata, features []Feature, origin Sequence) Record {
	record := &recordType{
		fields:   fields,
		features: make([]Feature, 0),

		origin: origin.Bytes(),
		insch:  make(chan insArg),
		delch:  make(chan delArg),
		repch:  make(chan repArg),
		locch:  make(chan Location),
		seqch:  make(chan Sequence),
	}
	for _, feature := range features {
		record.AddFeature(feature)
	}
	record.Start()
	return record
}

func (record *recordType) Start() {
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

func (record recordType) Fields() *Metadata {
	return record.fields
}

func (record recordType) Features() []Feature {
	return record.features
}

func (record *recordType) AddFeature(feature Feature) {
	if f, ok := feature.(*featureType); ok {
		f.insch = record.insch
		f.delch = record.delch
		f.repch = record.repch
		f.locch = record.locch
		f.seqch = record.seqch
	}
	i := sort.Search(len(record.features), func(i int) bool {
		compare := record.features[i]
		if compare.Key() == "source" && feature.Key() != "source" {
			return false
		}
		if feature.Key() == "source" && compare.Key() != "source" {
			return true
		}
		return LocationSmaller(feature.Location(), compare.Location())
	})
	features := make([]Feature, len(record.features)+1)
	copy(features[:i], record.features[:i])
	copy(features[i+1:], record.features[i:])
	features[i] = feature
	record.features = features
}

func (record recordType) Bytes() []byte {
	return record.origin
}

func (record recordType) String() string {
	return string(record.origin)
}

func (record recordType) Length() int {
	return len(record.origin)
}

func (record recordType) Slice(start, end int) Sequence {
	for start < 0 {
		start += len(record.origin)
	}
	for end < 0 {
		end += len(record.origin)
	}
	return Seq(record.origin[start:end])
}

func (record recordType) Subseq(loc Location) Sequence {
	return loc.Locate(record)
}

func (record *recordType) Insert(pos int, seq Sequence) {
	if seq.Length() == 0 {
		return
	}

	record.origin = insertBytes(record.origin, pos, seq.Bytes())
	for _, feature := range record.Features() {
		feature.Location().Shift(pos, seq.Length())
	}
}

func (record *recordType) Delete(pos, cnt int) {
	if cnt == 0 {
		return
	}

	record.origin = deleteBytes(record.origin, pos, cnt)
	for _, feature := range record.Features() {
		feature.Location().Shift(pos+cnt, -cnt)
	}
}

func (record *recordType) Replace(pos int, seq Sequence) {
	if seq.Length() == 0 {
		return
	}

	record.origin = replaceBytes(record.origin, pos, seq.Bytes())
}

var RecordParser = pars.Any(GenBankParser)

func ReadRecord(r io.Reader) (Record, error) {
	state := pars.NewState(r)
	result, err := pars.Apply(RecordParser, state)
	if err != nil {
		return nil, err
	}
	return result.(Record), nil
}
