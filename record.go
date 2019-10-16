package gt1

import (
	"time"

	"github.com/ktnyt/gods"
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
	DBLink     *gods.Ordered
	Definition string
	Keywords   []string
	Source     Organism
	References []Reference
	Comment    string
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

type Record struct {
	metadata *Metadata
	features *FeatureTable
	sequence []byte

	insch chan insArg
	delch chan delArg
	repch chan repArg
	locch chan Location
	seqch chan Sequence
}

func NewRecord(metadata *Metadata, features *FeatureTable, sequence BytesLike) *Record {
	record := &Record{
		metadata: metadata,
		features: features,
		sequence: AsBytes(sequence),

		insch: make(chan insArg),
		delch: make(chan delArg),
		repch: make(chan repArg),
		locch: make(chan Location),
		seqch: make(chan Sequence),
	}
	record.Start()
	return record
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

func (record Record) Metadata() *Metadata {
	return record.metadata
}

func (record Record) Features() *FeatureTable {
	return record.features
}

func (record *Record) AddFeature(feature *Feature) {
	feature.insch = record.insch
	feature.delch = record.delch
	feature.repch = record.repch
	feature.locch = record.locch
	feature.seqch = record.seqch
	record.features.Add(feature)
}

func (record Record) Bytes() []byte {
	return record.sequence
}

func (record Record) String() string {
	return string(record.sequence)
}

func (record Record) Len() int {
	return len(record.sequence)
}

func (record Record) Slice(start, end int) Sequence {
	for start < 0 {
		start += len(record.sequence)
	}
	for end < 0 {
		end += len(record.sequence)
	}
	return Seq(record.sequence[start:end])
}

func (record Record) Subseq(loc Location) Sequence {
	return loc.Locate(record)
}

func (record *Record) Insert(pos int, seq Sequence) {
	if seq.Len() == 0 {
		return
	}

	record.sequence = insertBytes(record.sequence, pos, seq.Bytes())
	for _, feature := range record.Features().Iter() {
		feature.Location().Shift(pos, seq.Len())
	}
}

func (record *Record) Delete(pos, cnt int) {
	if cnt == 0 {
		return
	}

	record.sequence = deleteBytes(record.sequence, pos, cnt)
	for _, feature := range record.Features().Iter() {
		feature.Location().Shift(pos+cnt, -cnt)
	}
}

func (record *Record) Replace(pos int, seq Sequence) {
	if seq.Len() == 0 {
		return
	}

	record.sequence = replaceBytes(record.sequence, pos, seq.Bytes())
}
