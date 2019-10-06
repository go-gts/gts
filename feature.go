package gt1

type Feature interface {
	Key() string
	Location() Location
	Qualifiers() *PairList
	Sequence
	Insert(pos int, seq Sequence)
	Delete(pos, cnt int)
	Replace(pos int, seq Sequence)
}

type featureType struct {
	key string
	loc Location
	qfs *PairList

	insch chan insArg
	delch chan delArg
	repch chan repArg
	locch chan Location
	seqch chan Sequence
}

func NewFeature(key string, loc Location, qfs *PairList) Feature {
	return &featureType{key: key, loc: loc, qfs: qfs}
}

func (feature featureType) Key() string           { return feature.key }
func (feature featureType) Location() Location    { return feature.loc }
func (feature featureType) Qualifiers() *PairList { return feature.qfs }

func (feature featureType) Seq() Sequence {
	if feature.locch == nil || feature.seqch == nil {
		panic("feature is not associated to a record: sequence is unavailable")
	}
	feature.locch <- feature.loc
	return <-feature.seqch
}

func (feature featureType) Bytes() []byte  { return feature.Seq().Bytes() }
func (feature featureType) String() string { return feature.Seq().String() }
func (feature featureType) Length() int    { return feature.Seq().Length() }

func (feature featureType) Slice(start, end int) Sequence {
	return feature.Seq().Slice(start, end)
}

func (feature featureType) Subseq(loc Location) Sequence {
	return feature.Seq().Subseq(loc)
}

func (feature featureType) Insert(pos int, seq Sequence) {
	if seq.Length() == 0 {
		return
	}

	if feature.insch != nil {
		panic("feature is not associated to a record: cannot insert sequence")
	}
	feature.insch <- insArg{feature.loc.Map(pos), seq}
}

func (feature featureType) Delete(pos, cnt int) {
	if cnt == 0 {
		return
	}

	if feature.delch != nil {
		panic("feature is not associated to a record: cannot delete sequence")
	}

	// Create a list of mapped indices.
	maps := make([]int, cnt)
	for i := 0; i < cnt; i++ {
		maps[i] = feature.loc.Map(pos + i)
	}

	for i := 1; i < cnt; i++ {
		// If there is a non-contiguous region, delete it separately.
		if maps[i-1]+1 != maps[i] {
			feature.delch <- delArg{maps[0], i}
			feature.Delete(pos+i, cnt-i)
			return
		}
	}

	feature.delch <- delArg{pos, cnt}
}

func (feature featureType) Replace(pos int, seq Sequence) {
	if seq.Length() == 0 {
		return
	}

	if feature.repch != nil {
		panic("feature is not associated to a record: cannot replace sequence")
	}

	// Create a list of mapped indices.
	maps := make([]int, seq.Length())
	for i := 0; i < seq.Length(); i++ {
		maps[i] = feature.loc.Map(pos + i)
	}

	for i := 1; i < seq.Length(); i++ {
		// If there is a non-contiguous region, replace it separately.
		if maps[i-1]+1 != maps[i] {
			feature.repch <- repArg{maps[0], seq.Slice(0, i)}
			feature.Replace(pos+i, seq.Slice(i, -1))
			return
		}
	}

	feature.repch <- repArg{pos, seq}
}
