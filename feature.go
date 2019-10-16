package gt1

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/ktnyt/gods"
	"github.com/ktnyt/pars"
	yaml "gopkg.in/yaml.v2"
)

type Feature struct {
	key string
	loc Location
	qfs *gods.Ordered

	insch chan insArg
	delch chan delArg
	repch chan repArg
	locch chan Location
	seqch chan Sequence
}

func NewFeature(key string, loc Location, qfs *gods.Ordered) *Feature {
	return &Feature{key: key, loc: loc, qfs: qfs}
}

func (feature Feature) Key() string               { return feature.key }
func (feature Feature) Location() Location        { return feature.loc }
func (feature Feature) Qualifiers() *gods.Ordered { return feature.qfs }

func (feature Feature) Seq() Sequence {
	if feature.locch == nil || feature.seqch == nil {
		panic("feature is not associated to a record: sequence is unavailable")
	}
	feature.locch <- feature.loc
	return <-feature.seqch
}

func (feature Feature) Bytes() []byte  { return feature.Seq().Bytes() }
func (feature Feature) String() string { return feature.Seq().String() }
func (feature Feature) Len() int       { return feature.Seq().Len() }

func (feature Feature) Slice(start, end int) Sequence {
	return feature.Seq().Slice(start, end)
}

func (feature Feature) Subseq(loc Location) Sequence {
	return feature.Seq().Subseq(loc)
}

func (feature Feature) Insert(pos int, seq Sequence) {
	if seq.Len() == 0 {
		return
	}

	if feature.insch != nil {
		panic("feature is not associated to a record: cannot insert sequence")
	}
	feature.insch <- insArg{feature.loc.Map(pos), seq}
}

func (feature Feature) Delete(pos, cnt int) {
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

func (feature Feature) Replace(pos int, seq Sequence) {
	if seq.Len() == 0 {
		return
	}

	if feature.repch != nil {
		panic("feature is not associated to a record: cannot replace sequence")
	}

	// Create a list of mapped indices.
	maps := make([]int, seq.Len())
	for i := 0; i < seq.Len(); i++ {
		maps[i] = feature.loc.Map(pos + i)
	}

	for i := 1; i < seq.Len(); i++ {
		// If there is a non-contiguous region, replace it separately.
		if maps[i-1]+1 != maps[i] {
			feature.repch <- repArg{maps[0], seq.Slice(0, i)}
			feature.Replace(pos+i, seq.Slice(i, -1))
			return
		}
	}

	feature.repch <- repArg{pos, seq}
}

func CompareFeatures(a, b *Feature) bool {
	if a.Key() == "source" && b.Key() != "source" {
		return true
	}
	if b.Key() == "source" && a.Key() != "source" {
		return false
	}
	return LocationSmaller(a.Location(), b.Location())
}

type FeatureTable []*Feature

func NewFeatureTable() *FeatureTable {
	ft := new([]*Feature)
	*ft = make([]*Feature, 0)
	return (*FeatureTable)(ft)
}

func (ft FeatureTable) Len() int {
	return len(ft)
}

func (ft FeatureTable) Get(i int) *Feature {
	return ft[i]
}

func (ft *FeatureTable) Add(feature *Feature) {
	features := *ft
	comp := func(i int) bool { return CompareFeatures(feature, features[i]) }
	i := sort.Search(len(*ft), comp)
	ret := make([]*Feature, len(features)+1)
	copy(ret[:i], features[:i])
	copy(ret[i+1:], features[i:])
	ret[i] = feature
	*ft = FeatureTable(ret)
}

func (ft *FeatureTable) Append(feature *Feature) {
	features := *ft
	features = append(features, feature)
	*ft = FeatureTable(features)
}

func (ft *FeatureTable) Iter() []*Feature {
	return []*Feature(*ft)
}

type FeatureFilter func(*Feature) bool

func FeatureFilterOr(filters ...FeatureFilter) FeatureFilter {
	return func(feature *Feature) bool {
		for _, filter := range filters {
			if filter(feature) {
				return true
			}
		}
		return false
	}
}

func FeatureFilterAnd(filters ...FeatureFilter) FeatureFilter {
	return func(feature *Feature) bool {
		for _, filter := range filters {
			if !filter(feature) {
				return false
			}
		}
		return true
	}
}

func FeatureFilterInvert(filter FeatureFilter) FeatureFilter {
	return func(feature *Feature) bool {
		return !filter(feature)
	}
}

func baseFilter(feature *Feature) bool {
	return feature.Key() == "source"
}

func FeatureKeyFilter(keys []string) FeatureFilter {
	return func(feature *Feature) bool {
		if len(keys) == 0 {
			return true
		}
		for i := range keys {
			if feature.Key() == keys[i] {
				return true
			}
		}
		return false
	}
}

func ClearFeatures(features *FeatureTable) *FeatureTable {
	return FilterFeatures(features, func(feature *Feature) bool { return false })
}

func FilterFeatures(features *FeatureTable, filter FeatureFilter) *FeatureTable {
	f := FeatureFilterOr(baseFilter, filter)

	matches := make([]bool, features.Len())
	count := 0
	for i, feature := range features.Iter() {
		if f(feature) {
			matches[i] = true
			count++
		}
	}

	ret := make([]*Feature, count)
	j := 0
	for i := range features.Iter() {
		if matches[i] {
			ret[j] = features.Get(i)
			j++
		}
	}

	return (*FeatureTable)(&ret)
}

type featureName struct {
	Indent int
	Value  string
	Depth  int
}

var featureNameParser = pars.Seq(
	pars.Many(' '),
	pars.SnakeWord.Map(pars.CatByte),
	pars.Many(' '),
).Map(func(result *pars.Result) error {
	indent := len(result.Children[0].Children)
	value := result.Children[1].Value.(string)
	depth := indent + len(value) + len(result.Children[2].Children)
	result.Value = featureName{Indent: indent, Value: value, Depth: depth}
	result.Children = nil
	return nil
})

func featureBodyParser(key string, indent, depth int) pars.Parser {
	depthString := "\n" + strings.Repeat(" ", depth)

	return func(state *pars.State, result *pars.Result) error {
		// First line must be a range.
		if err := locationParser(state, result); err != nil {
			return err
		}
		loc := result.Value.(Location)
		pars.Try('\n')(state, result)

		qfs := gods.NewOrdered()

		for {
			// Count the leading spaces.
			state.Mark()

			count := 0

			if err := state.Want(1); err != nil {
				state.Jump()
				return err
			}
			for state.Buffer[state.Index] == ' ' {
				state.Advance(1)
				count += 1
				if err := state.Want(1); err != nil {
					state.Jump()
					return err
				}
			}

			// End of feature so return.
			if count <= indent {
				state.Jump()
				result.Value = NewFeature(key, loc, qfs)
				result.Children = nil
				return nil
			}

			if count != depth {
				state.Jump()
				return pars.NewMismatchError("GeBank Feature Field", []byte("matching depth"), state.Position)
			}

			// Remaining fields must be preceded by a /.
			if state.Buffer[state.Index] != '/' {
				state.Jump()
				return pars.NewMismatchError("Feature Field", []byte{'/'}, state.Position)
			}

			if err := state.Want(1); err != nil {
				state.Jump()
				return err
			}
			state.Advance(1)

			// Match the name of the feature field.
			if err := pars.Until(pars.Any('=', '\n'))(state, result); err != nil {
				state.Jump()
				return err
			}
			name := result.Value.(string)

			if state.Buffer[state.Index] == '\n' {
				pars.Try('\n')(state, result)
				qfs.Add(name, "")
				state.Unmark()
				continue
			}

			// Next byte is guaranteed to be = by Until.
			state.Advance(1)

			// In most cases the feature property values are quoted.
			if err := pars.Quoted('"')(state, result); err != nil {
				// Otherwise just get the rest of the line.
				if err := pars.Line(state, result); err != nil {
					state.Jump()
					return err
				}
			}
			value := result.Value.(string)

			// Completely remove indents for translations.
			if name == "translation" {
				value = strings.Replace(value, depthString, "", -1)
			} else {
				value = strings.Replace(value, depthString, " ", -1)
			}

			// Remove the newline.
			pars.Try('\n')(state, result)

			qfs.Add(name, value)

			state.Unmark()
		}
	}
}

func FeatureTableParser(state *pars.State, result *pars.Result) error {
	// Discard the Location/Qualifiers line.
	if err := pars.Line(state, result); err != nil {
		return pars.NewTraceError("Feature", err)
	}

	if err := featureNameParser(state, result); err != nil {
		return pars.NewTraceError("Feature", err)
	}

	name := result.Value.(featureName)

	features := NewFeatureTable()

	key := name.Value
	indent := name.Indent
	depth := name.Depth

	// Process the source feature body.
	if err := featureBodyParser(key, indent, depth)(state, result); err != nil {
		return pars.NewTraceError("Feature", err)
	}

	features.Append(result.Value.(*Feature))

	// Continually process feature properties while indented.
	for state.Buffer[state.Index] == ' ' {
		if err := featureNameParser(state, result); err != nil {
			return pars.NewTraceError("Feature", err)
		}
		key = result.Value.(featureName).Value

		if err := featureBodyParser(key, indent, depth)(state, result); err != nil {
			return pars.NewTraceError("Feature", err)
		}
		features.Append(result.Value.(*Feature))
	}

	result.Value = features
	result.Children = nil
	return nil
}

func recordToFeatureTable(result *pars.Result) error {
	switch v := result.Value.(type) {
	case Record:
		result.Value = v.Features()
		result.Children = nil
		return nil
	default:
		return fmt.Errorf("cannot convert type `%T` to Record", v)
	}
}

type featureIO struct {
	Key        *string
	Location   *string
	Qualifiers [][]string
}

func featureTableFromFeatureIOSlice(fios []featureIO) (*FeatureTable, error) {
	features := NewFeatureTable()
	for i, fio := range fios {
		if fio.Key == nil {
			return nil, fmt.Errorf("%s feature is missing a key", humanize.Ordinal(i+1))
		}
		if fio.Location == nil {
			return nil, fmt.Errorf("%s feature is missing a location", humanize.Ordinal(i+1))
		}
		qfs := gods.NewOrdered()
		for _, item := range fio.Qualifiers {
			qfs.Add(item[0], item[1])
		}
		features.Append(NewFeature(*(fio.Key), AsLocation(*(fio.Location)), qfs))
	}
	return features, nil
}

func featureYamlParser(state *pars.State, result *pars.Result) error {
	pSlice := new([]featureIO)
	decoder := yaml.NewDecoder(state)
	state.Mark()
	if err := decoder.Decode(pSlice); err != nil {
		state.Jump()
		return err
	}
	state.Unmark()

	features, err := featureTableFromFeatureIOSlice(*pSlice)
	if err != nil {
		return err
	}
	result.Value = features
	result.Children = nil
	return nil
}

func featureJsonParser(state *pars.State, result *pars.Result) error {
	pSlice := new([]featureIO)
	decoder := json.NewDecoder(state)
	state.Mark()
	if err := decoder.Decode(pSlice); err != nil {
		state.Jump()
		return err
	}

	state.Unmark()
	features, err := featureTableFromFeatureIOSlice(*pSlice)
	if err != nil {
		return err
	}
	result.Value = features
	result.Children = nil
	return nil
}

var FeatureParser = pars.Any(
	featureYamlParser,
	featureJsonParser,
)

func ReadFeatures(r io.Reader) (*FeatureTable, error) {
	state := pars.NewState(r)
	result, err := pars.Apply(FeatureParser, state)
	if err != nil {
		switch e := err.(type) {
		case *pars.TraceError:
			return nil, e.Unwrap()
		default:
			return nil, e
		}
		return nil, err
	}
	return result.(*FeatureTable), nil
}
