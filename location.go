package gt1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ktnyt/pars"
)

// Location represents a feature location as defined by the INSDC.
type Location interface {
	// Locate the sequence at the pointing location.
	Locate(seq Sequence) Sequence

	// Len returns the length spanned by the location.
	Len() int

	// Format the location.
	Format() string

	// Shift the location position[s] if needed.
	Shift(pos, n int)

	// Map the given local index to a global index.
	Map(index int) int
}

// LocationSmaller tests if the one location is smaller than the other.
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

// PointLocation represents a single point.
type PointLocation struct{ Position int }

// NewPointLocation creates a new PointLocation.
func NewPointLocation(pos int) Location {
	return &PointLocation{Position: pos}
}

// Locate the sequence at the pointing location.
func (location PointLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Position, location.Position+1)
}

// Len returns the length spanned by the location.
func (location PointLocation) Len() int {
	return 1
}

// Format the location.
func (location PointLocation) Format() string {
	return strconv.Itoa(location.Position + 1)
}

// Shift the location position[s] if needed.
func (location *PointLocation) Shift(pos, n int) {
	if pos <= location.Position {
		location.Position += n
	}
}

// Map the given local index to a global index.
func (location PointLocation) Map(index int) int {
	index = fixIndex(index, location.Len())
	return location.Position + index
}

// RangeLocation represents a range of locations.
type RangeLocation struct {
	Start    int
	End      int
	Partial5 bool
	Partial3 bool
}

// NewRangeLocation creates a new RangeLocation.
func NewRangeLocation(start, end int) Location {
	return &RangeLocation{Start: start, End: end}
}

// NewPartialRangeLocation creates a new partial RangeLocation.
func NewPartialRangeLocation(start, end int, p5, p3 bool) Location {
	return &RangeLocation{
		Start:    start,
		End:      end,
		Partial5: p5,
		Partial3: p3,
	}
}

// Locate the sequence at the pointing location.
func (location RangeLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Start, location.End)
}

// Len returns the length spanned by the location.
func (location RangeLocation) Len() int {
	return location.End - location.Start
}

// Format the location.
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

// Shift the location position[s] if needed.
func (location *RangeLocation) Shift(pos, n int) {
	if pos <= location.Start {
		location.Start += n
	}
	if pos <= location.End {
		location.End += n
	}
}

// Map the given local index to a global index.
func (location RangeLocation) Map(index int) int {
	index = fixIndex(index, location.Len())
	return location.Start + index
}

// AmbiguousLocation represents an ambiguous location.
type AmbiguousLocation struct {
	Start int
	End   int
}

// NewAmbiguousLocation creates a new ambiguous location.
func NewAmbiguousLocation(start, end int) Location {
	return &AmbiguousLocation{Start: start, End: end}
}

// Locate the sequence at the pointing location.
func (location AmbiguousLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Start, location.End)
}

// Len returns the length spanned by the location.
func (location AmbiguousLocation) Len() int {
	return location.End - location.Start
}

// Format the location.
func (location AmbiguousLocation) Format() string {
	return fmt.Sprintf("%d.%d", location.Start+1, location.End)
}

// Shift the location position[s] if needed.
func (location *AmbiguousLocation) Shift(pos, n int) {
	if pos <= location.Start {
		location.Start += n
	}
	if pos <= location.End {
		location.End += n
	}
}

// Map the given local index to a global index.
func (location AmbiguousLocation) Map(index int) int {
	index = fixIndex(index, location.Len())
	return location.Start + index
}

// BetweenLocation represents a location between two points.
type BetweenLocation struct {
	Start int
	End   int
}

// NewBetweenLocation creates a new BetweenLocation.
func NewBetweenLocation(start, end int) Location {
	return &BetweenLocation{Start: start, End: end}
}

// Locate the sequence at the pointing location.
func (location BetweenLocation) Locate(seq Sequence) Sequence {
	return seq.Slice(location.Start, location.End)
}

// Len returns the length spanned by the location.
func (location BetweenLocation) Len() int {
	return location.End - location.Start
}

// Format the location.
func (location BetweenLocation) Format() string {
	return fmt.Sprintf("%d^%d", location.Start+1, location.End)
}

// Shift the location position[s] if needed.
func (location *BetweenLocation) Shift(pos, n int) {
	if pos <= location.Start {
		location.Start += n
	}
	if pos <= location.End {
		location.End += n
	}
}

// Map the given local index to a global index.
func (location BetweenLocation) Map(index int) int {
	index = fixIndex(index, location.Len())
	return location.Start + index
}

// ComplementLocation represents the complement region of a location.
type ComplementLocation struct {
	Location Location
}

// NewComplementLocation creates a new ComplementLocation.
func NewComplementLocation(location Location) Location {
	return &ComplementLocation{Location: location}
}

// Locate the sequence at the pointing location.
func (location ComplementLocation) Locate(seq Sequence) Sequence {
	return Complement(seq.Subseq(location.Location))
}

// Len returns the length spanned by the location.
func (location ComplementLocation) Len() int {
	return location.Location.Len()
}

// Format the location.
func (location ComplementLocation) Format() string {
	return fmt.Sprintf("complement(%s)", location.Location.Format())
}

// Shift the location position[s] if needed.
func (location *ComplementLocation) Shift(pos, n int) {
	location.Location.Shift(pos, n)
}

// Map the given local index to a global index.
func (location ComplementLocation) Map(index int) int {
	return location.Location.Map(index)
}

// JoinLocation represents multiple joined locations.
type JoinLocation struct {
	Locations []Location
}

// NewJoinLocation creates a new JoinLocation.
func NewJoinLocation(locations []Location) Location {
	return &JoinLocation{Locations: locations}
}

// Locate the sequence at the pointing location.
func (location JoinLocation) Locate(seq Sequence) Sequence {
	r := make([]byte, location.Len())
	i := 0
	for _, l := range location.Locations {
		copy(r[i:], l.Locate(seq).Bytes())
		i += l.Len()
	}
	return Seq(r)
}

// Len returns the length spanned by the location.
func (location JoinLocation) Len() int {
	length := 0
	for _, l := range location.Locations {
		length += l.Len()
	}
	return length
}

// Format the location.
func (location JoinLocation) Format() string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Format()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

// Shift the location position[s] if needed.
func (location *JoinLocation) Shift(pos, n int) {
	for i := range location.Locations {
		location.Locations[i].Shift(pos, n)
	}
}

// Map the given local index to a global index.
func (location JoinLocation) Map(index int) int {
	index = fixIndex(index, location.Len())
	for _, l := range location.Locations {
		if index < l.Len() {
			return l.Map(index)
		}
		index -= l.Len()
	}
	panic("the program should never reach this state...")
}

// OrderLocation represents a group of locations.
type OrderLocation struct {
	Locations []Location
}

// NewOrderLocation creates a new OrderLocation.
func NewOrderLocation(locations []Location) Location {
	return &OrderLocation{Locations: locations}
}

// Locate the sequence at the pointing location.
func (location OrderLocation) Locate(seq Sequence) Sequence {
	r := make([]byte, location.Len())
	i := 0
	for _, l := range location.Locations {
		copy(r[i:], l.Locate(seq).Bytes())
		i += l.Len()
	}
	return Seq(r)
}

// Len returns the length spanned by the location.
func (location OrderLocation) Len() int {
	length := 0
	for _, l := range location.Locations {
		length += l.Len()
	}
	return length
}

// Format the location.
func (location OrderLocation) Format() string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Format()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

// Shift the location position[s] if needed.
func (location *OrderLocation) Shift(pos, n int) {
	for i := range location.Locations {
		location.Locations[i].Shift(pos, n)
	}
}

// Map the given local index to a global index.
func (location OrderLocation) Map(index int) int {
	index = fixIndex(index, location.Len())
	for _, l := range location.Locations {
		if index < l.Len() {
			return l.Map(index)
		}
		index -= l.Len()
	}
	panic("the program should never reach this state...")
}

var LocationParser pars.Parser

// PointLocationParser attempts to parse a PointLocation.
var PointLocationParser = pars.Parser(pars.Int).Map(func(result *pars.Result) error {
	n := result.Value.(int)
	loc := NewPointLocation(n - 1)
	result.SetValue(loc)
	return nil
})

// RangeLocationParser attempts to parse a RangeLocation.
var RangeLocationParser = pars.Seq(
	pars.Maybe('<'), pars.Int, "..", pars.Maybe('>'), pars.Int,
	pars.Maybe('>'), // Possibly required for some legacy entries.
).Map(func(result *pars.Result) error {
	loc := NewPartialRangeLocation(
		result.Children[1].Value.(int)-1,
		result.Children[4].Value.(int),
		result.Children[0].Value != nil,
		result.Children[3].Value != nil || result.Children[5].Value != nil,
	)
	result.SetValue(loc)
	return nil
})

// AmbiguousLocationParser attempts to parse a AmbiguousLocation.
var AmbiguousLocationParser = pars.Seq(
	pars.Int, '.', pars.Int,
).Map(func(result *pars.Result) error {
	loc := NewAmbiguousLocation(
		result.Children[0].Value.(int)-1,
		result.Children[2].Value.(int),
	)
	result.SetValue(loc)
	return nil
})

// BetweenLocationParser attempts to parse a BetweenLocation.
var BetweenLocationParser = pars.Seq(
	pars.Int, '^', pars.Int,
).Map(func(result *pars.Result) error {
	loc := NewBetweenLocation(
		result.Children[0].Value.(int)-1,
		result.Children[2].Value.(int),
	)
	result.SetValue(loc)
	return nil
})

// ComplementLocationParser attempts to parse a ComplementLocation.
var ComplementLocationParser = pars.Seq(
	"complement(", &LocationParser, ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	loc := NewComplementLocation(result.Value.(Location))
	result.SetValue(loc)
	return nil
})

var locationDelimiter = pars.Seq(',', pars.Many(pars.Space))

// JoinLocationParser attempts to parse a JoinLocation.
var JoinLocationParser = pars.Seq(
	"join(", pars.Delim(&LocationParser, locationDelimiter), ')',
).Map(func(result *pars.Result) error {
	children := result.Children[1].Children
	locations := make([]Location, len(children))
	for i, child := range children {
		locations[i] = child.Value.(Location)
	}
	loc := NewJoinLocation(locations)
	result.SetValue(loc)
	return nil
})

// OrderLocationParser attempts to parse a OrderLocation.
var OrderLocationParser = pars.Seq(
	"order(", pars.Delim(&LocationParser, locationDelimiter), ')',
).Map(func(result *pars.Result) error {
	children := result.Children[1].Children
	locations := make([]Location, len(children))
	for i, child := range children {
		locations[i] = child.Value.(Location)
	}
	loc := NewOrderLocation(locations)
	result.SetValue(loc)
	return nil
})

// ParseLocation attempts to parse any Location.
func ParseLocation(s string) Location {
	state := pars.FromString(s)
	result, err := LocationParser.Parse(state)
	if err != nil {
		panic("could not interpret string as Location")
	}
	return result.Value.(Location)
}

func init() {
	LocationParser = pars.Any(
		RangeLocationParser,
		OrderLocationParser,
		JoinLocationParser,
		ComplementLocationParser,
		AmbiguousLocationParser,
		BetweenLocationParser,
		PointLocationParser,
	)
}
