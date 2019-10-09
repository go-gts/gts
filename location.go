package gt1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ktnyt/pars"
)

type Location interface {
	// Locate the bytes at the pointed location.
	Locate(seq Sequence) Sequence

	// Length returns the length of the pointed location.
	Length() int

	// Format the location.
	Format() string

	// Shift the location if needed.
	Shift(pos, n int)

	// Map the given local index to a global index.
	Map(index int) int
}

func LocationSmaller(a, b Location) bool {
	if a.Map(0) < b.Map(0) {
		return true
	}
	if b.Map(0) < a.Map(0) {
		return false
	}
	if a.Map(-1) < b.Map(-1) {
		return true
	}
	if b.Map(-1) < a.Map(-1) {
		return false
	}
	return false
}

func fixIndex(index, length int) int {
	if index < 0 {
		index += length
	}
	if index >= length {
		panic(fmt.Errorf("index [%d] is outside of location with length %d", index, length))
	}
	return index
}

type PointLocation struct {
	Position int
}

func NewPointLocation(pos int) Location {
	return &PointLocation{Position: pos}
}

func (location PointLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Position, location.Position+1)
}

func (location PointLocation) Length() int {
	return 1
}

func (location PointLocation) Format() string {
	return strconv.Itoa(location.Position + 1)
}

func (location *PointLocation) Shift(pos, n int) {
	if pos <= location.Position {
		location.Position += n
	}
}

func (location PointLocation) Map(index int) int {
	index = fixIndex(index, location.Length())
	return location.Position + index
}

type RangeLocation struct {
	Start    int
	End      int
	Partial5 bool
	Partial3 bool
}

func NewRangeLocation(start, end int) Location {
	return &RangeLocation{Start: start, End: end}
}

func NewPartialRangeLocation(start, end int, p5, p3 bool) Location {
	return &RangeLocation{
		Start:    start,
		End:      end,
		Partial5: p5,
		Partial3: p3,
	}
}

func (location RangeLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Start, location.End)
}

func (location RangeLocation) Length() int {
	return location.End - location.Start
}

func (location RangeLocation) Format() string {
	p5, p3 := "", ""
	if location.Partial5 {
		p5 = "<"
	}
	if location.Partial3 {
		p3 = ">"
	}
	return fmt.Sprintf("%s%d..%s%d", p5, location.Start+1, p3, location.End)
}

func (location *RangeLocation) Shift(pos, n int) {
	if pos <= location.Start {
		location.Start += n
	}
	if pos <= location.End {
		location.End += n
	}
}

func (location RangeLocation) Map(index int) int {
	index = fixIndex(index, location.Length())
	return location.Start + index
}

type AmbiguousLocation struct {
	Start int
	End   int
}

func NewAmbiguousLocation(start, end int) Location {
	return &AmbiguousLocation{Start: start, End: end}
}

func (location AmbiguousLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Start, location.End)
}

func (location AmbiguousLocation) Length() int {
	return location.End - location.Start
}

func (location AmbiguousLocation) Format() string {
	return fmt.Sprintf("%d.%d", location.Start+1, location.End)
}

func (location *AmbiguousLocation) Shift(pos, n int) {
	if pos <= location.Start {
		location.Start += n
	}
	if pos <= location.End {
		location.End += n
	}
}

func (location AmbiguousLocation) Map(index int) int {
	index = fixIndex(index, location.Length())
	return location.Start + index
}

type BetweenLocation struct {
	Start int
	End   int
}

func NewBetweenLocation(start, end int) Location {
	return &BetweenLocation{Start: start, End: end}
}

func (location BetweenLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Start, location.End)
}

func (location BetweenLocation) Length() int {
	return location.End - location.Start
}

func (location BetweenLocation) Format() string {
	return fmt.Sprintf("%d^%d", location.Start+1, location.End)
}

func (location *BetweenLocation) Shift(pos, n int) {
	if pos <= location.Start {
		location.Start += n
	}
	if pos <= location.End {
		location.End += n
	}
}

func (location BetweenLocation) Map(index int) int {
	index = fixIndex(index, location.Length())
	return location.Start + index
}

type ComplementLocation struct {
	Location Location
}

func NewComplementLocation(location Location) Location {
	return &ComplementLocation{Location: location}
}

func (location ComplementLocation) Locate(seq Sequence) Sequence {
	return Complement(seq.Subseq(location.Location))
}

func (location ComplementLocation) Length() int {
	return location.Location.Length()
}

func (location ComplementLocation) Format() string {
	return fmt.Sprintf("complement(%s)", location.Location.Format())
}

func (location *ComplementLocation) Shift(pos, n int) {
	location.Location.Shift(pos, n)
}

func (location ComplementLocation) Map(index int) int {
	return location.Location.Map(index)
}

type JoinLocation struct {
	Locations []Location
}

func NewJoinLocation(locations []Location) Location {
	return &JoinLocation{Locations: locations}
}

func (location JoinLocation) Locate(seq Sequence) Sequence {
	r := make([]byte, location.Length())
	i := 0
	for _, l := range location.Locations {
		copy(r[i:], l.Locate(seq).Bytes())
		i += l.Length()
	}
	return Seq(r)
}

func (location JoinLocation) Length() int {
	length := 0
	for _, l := range location.Locations {
		length += l.Length()
	}
	return length
}

func (location JoinLocation) Format() string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Format()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

func (location *JoinLocation) Shift(pos, n int) {
	for i := range location.Locations {
		location.Locations[i].Shift(pos, n)
	}
}

func (location JoinLocation) Map(index int) int {
	index = fixIndex(index, location.Length())
	for _, l := range location.Locations {
		if index < l.Length() {
			return l.Map(index)
		}
		index -= l.Length()
	}
	panic("the program should never reach this state...")
}

type OrderLocation struct {
	Locations []Location
}

func NewOrderLocation(locations []Location) Location {
	return &OrderLocation{Locations: locations}
}

func (location OrderLocation) Locate(seq Sequence) Sequence {
	r := make([]byte, location.Length())
	i := 0
	for _, l := range location.Locations {
		copy(r[i:], l.Locate(seq).Bytes())
		i += l.Length()
	}
	return Seq(r)
}
func (location OrderLocation) Length() int {
	length := 0
	for _, l := range location.Locations {
		length += l.Length()
	}
	return length
}

func (location OrderLocation) Format() string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Format()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

func (location *OrderLocation) Shift(pos, n int) {
	for i := range location.Locations {
		location.Locations[i].Shift(pos, n)
	}
}

func (location OrderLocation) Map(index int) int {
	index = fixIndex(index, location.Length())
	for _, l := range location.Locations {
		if index < l.Length() {
			return l.Map(index)
		}
		index -= l.Length()
	}
	panic("the program should never reach this state...")
}

var locationParser pars.Parser

var pointLocationParser = pars.Integer.Map(func(result *pars.Result) error {
	n, err := strconv.Atoi(result.Value.(string))
	if err != nil {
		return err
	}
	result.Value = NewPointLocation(n - 1)
	result.Children = nil
	return nil
})

var rangeLocationParser = pars.Seq(
	pars.Try('<'),
	pars.Integer.Map(pars.Atoi),
	"..",
	pars.Try('>'),
	pars.Integer.Map(pars.Atoi),
	pars.Try('>'), // Possibly required for some legacy entries.
).Map(func(result *pars.Result) error {
	result.Value = NewPartialRangeLocation(
		result.Children[1].Value.(int)-1,
		result.Children[4].Value.(int),
		result.Children[0].Value != nil,
		result.Children[3].Value != nil || result.Children[5].Value != nil,
	)
	result.Children = nil
	return nil
})

var ambiguousLocationParser = pars.Seq(
	pars.Integer.Map(pars.Atoi), '.', pars.Integer.Map(pars.Atoi),
).Map(func(result *pars.Result) error {
	result.Value = NewAmbiguousLocation(
		result.Children[0].Value.(int)-1,
		result.Children[2].Value.(int),
	)
	result.Children = nil
	return nil
})

var betweenLocationParser = pars.Seq(
	pars.Integer.Map(pars.Atoi), '^', pars.Integer.Map(pars.Atoi),
).Map(func(result *pars.Result) error {
	result.Value = NewBetweenLocation(
		result.Children[0].Value.(int)-1,
		result.Children[2].Value.(int),
	)
	result.Children = nil
	return nil
})

var complementLocationParser = pars.Seq(
	"complement(", &locationParser, ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	result.Value = NewComplementLocation(result.Value.(Location))
	result.Children = nil
	return nil
})

var joinLocationParser = pars.Seq(
	"join(", pars.Delim(&locationParser, ','), ')',
).Map(func(result *pars.Result) error {
	children := result.Children[1].Children
	locations := make([]Location, len(children))
	for i, child := range children {
		locations[i] = child.Value.(Location)
	}
	result.Value = NewJoinLocation(locations)
	result.Children = nil
	return nil
})

var orderLocationParser = pars.Seq(
	"order(", pars.Delim(&locationParser, ','), ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	locations := make([]Location, len(result.Children))
	for i, child := range result.Children {
		locations[i] = child.Value.(Location)
	}
	result.Value = NewOrderLocation(locations)
	result.Children = nil
	return nil
})

func AsLocation(s string) Location {
	state := pars.FromString(s)
	result, err := pars.Apply(locationParser, state)
	if err != nil {
		panic("could not interpret string as Location")
	}
	return result.(Location)
}

func init() {
	locationParser = pars.Any(
		rangeLocationParser,
		orderLocationParser,
		joinLocationParser,
		complementLocationParser,
		ambiguousLocationParser,
		betweenLocationParser,
		pointLocationParser,
	)
}
