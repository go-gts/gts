package gt1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ktnyt/pars"
)

// Location represents a feature location as defined by the INSDC.
type Location interface {
	// Map the sequence at the pointing loc.
	Map(seq Sequence) Sequence

	// Len returns the length spanned by the loc.
	Len() int

	// Format the loc.
	Format() string

	// Shift the loc position[s] if needed.
	Shift(pos, n int)

	// Convert the given local index to a global index.
	Convert(index int) int
}

// LocationSmaller tests if the one location is smaller than the other.
func LocationSmaller(a, b Location) bool {
	if a.Convert(0) < b.Convert(0) {
		return true
	}
	if b.Convert(0) < a.Convert(0) {
		return false
	}
	if a.Convert(-1) < b.Convert(-1) {
		return true
	}
	if b.Convert(-1) < a.Convert(-1) {
		return false
	}
	return false
}

func fixIndex(index, length int) int {
	if index < 0 {
		index += length
	}
	if index >= length {
		panic(fmt.Errorf("index [%d] is outside of loc with length %d", index, length))
	}
	return index
}

// PointLocation represents a single point.
type PointLocation struct{ Position int }

// NewPointLocation creates a new PointLocation.
func NewPointLocation(pos int) Location {
	return &PointLocation{Position: pos}
}

// Map the sequence at the pointing location.
func (loc PointLocation) Map(seq Sequence) Sequence {
	return Slice(seq, loc.Position, loc.Position+1)
}

// Len returns the length spanned by the location.
func (loc PointLocation) Len() int {
	return 1
}

// Format the location.
func (loc PointLocation) Format() string {
	return strconv.Itoa(loc.Position + 1)
}

// Shift the location position[s] if needed.
func (loc *PointLocation) Shift(pos, n int) {
	if pos <= loc.Position {
		loc.Position += n
	}
}

// Convert the given local index to a global index.
func (loc PointLocation) Convert(index int) int {
	index = fixIndex(index, loc.Len())
	return loc.Position + index
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

// Map the sequence at the pointing location.
func (loc RangeLocation) Map(seq Sequence) Sequence {
	return Slice(seq, loc.Start, loc.End)
}

// Len returns the length spanned by the location.
func (loc RangeLocation) Len() int {
	return loc.End - loc.Start
}

// Format the location.
func (loc RangeLocation) Format() string {
	p5, p3 := "", ""
	if loc.Partial5 {
		p5 = "<"
	}
	if loc.Partial3 {
		p3 = ">"
	}
	return fmt.Sprintf("%s%d..%s%d", p5, loc.Start+1, p3, loc.End)
}

// Shift the location position[s] if needed.
func (loc *RangeLocation) Shift(pos, n int) {
	if pos <= loc.Start {
		loc.Start += n
	}
	if pos <= loc.End {
		loc.End += n
	}
}

// Convert the given local index to a global index.
func (loc RangeLocation) Convert(index int) int {
	index = fixIndex(index, loc.Len())
	return loc.Start + index
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

// Map the sequence at the pointing location.
func (loc AmbiguousLocation) Map(seq Sequence) Sequence {
	return Slice(seq, loc.Start, loc.End)
}

// Len returns the length spanned by the location.
func (loc AmbiguousLocation) Len() int {
	return loc.End - loc.Start
}

// Format the location.
func (loc AmbiguousLocation) Format() string {
	return fmt.Sprintf("%d.%d", loc.Start+1, loc.End)
}

// Shift the location position[s] if needed.
func (loc *AmbiguousLocation) Shift(pos, n int) {
	if pos <= loc.Start {
		loc.Start += n
	}
	if pos <= loc.End {
		loc.End += n
	}
}

// Convert the given local index to a global index.
func (loc AmbiguousLocation) Convert(index int) int {
	index = fixIndex(index, loc.Len())
	return loc.Start + index
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

// Map the sequence at the pointing location.
func (loc BetweenLocation) Map(seq Sequence) Sequence {
	return Slice(seq, loc.Start, loc.End)
}

// Len returns the length spanned by the location.
func (loc BetweenLocation) Len() int {
	return loc.End - loc.Start
}

// Format the location.
func (loc BetweenLocation) Format() string {
	return fmt.Sprintf("%d^%d", loc.Start+1, loc.End)
}

// Shift the location position[s] if needed.
func (loc *BetweenLocation) Shift(pos, n int) {
	if pos <= loc.Start {
		loc.Start += n
	}
	if pos <= loc.End {
		loc.End += n
	}
}

// Convert the given local index to a global index.
func (loc BetweenLocation) Convert(index int) int {
	index = fixIndex(index, loc.Len())
	return loc.Start + index
}

// ComplementLocation represents the complement region of a location.
type ComplementLocation struct {
	Location Location
}

// NewComplementLocation creates a new ComplementLocation.
func NewComplementLocation(loc Location) Location {
	return &ComplementLocation{Location: loc}
}

// Map the sequence at the pointing location.
func (loc ComplementLocation) Map(seq Sequence) Sequence {
	return Complement(loc.Location.Map(seq))
}

// Len returns the length spanned by the location.
func (loc ComplementLocation) Len() int {
	return loc.Location.Len()
}

// Format the location.
func (loc ComplementLocation) Format() string {
	return fmt.Sprintf("complement(%s)", loc.Location.Format())
}

// Shift the location position[s] if needed.
func (loc *ComplementLocation) Shift(pos, n int) {
	loc.Location.Shift(pos, n)
}

// Convert the given local index to a global index.
func (loc ComplementLocation) Convert(index int) int {
	return loc.Location.Convert(index)
}

// JoinLocation represents multiple joined locations.
type JoinLocation struct {
	Locations []Location
}

// NewJoinLocation creates a new JoinLocation.
func NewJoinLocation(locs []Location) Location {
	return &JoinLocation{Locations: locs}
}

// Map the sequence at the pointing location.
func (loc JoinLocation) Map(seq Sequence) Sequence {
	r := make([]byte, loc.Len())
	i := 0
	for _, l := range loc.Locations {
		copy(r[i:], l.Map(seq).Bytes())
		i += l.Len()
	}
	return Seq(r)
}

// Len returns the length spanned by the location.
func (loc JoinLocation) Len() int {
	length := 0
	for _, l := range loc.Locations {
		length += l.Len()
	}
	return length
}

// Format the location.
func (loc JoinLocation) Format() string {
	tmp := make([]string, len(loc.Locations))
	for i := range loc.Locations {
		tmp[i] = loc.Locations[i].Format()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

// Shift the location position[s] if needed.
func (loc *JoinLocation) Shift(pos, n int) {
	for i := range loc.Locations {
		loc.Locations[i].Shift(pos, n)
	}
}

// Convert the given local index to a global index.
func (loc JoinLocation) Convert(index int) int {
	index = fixIndex(index, loc.Len())
	for _, l := range loc.Locations {
		if index < l.Len() {
			return l.Convert(index)
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
func NewOrderLocation(locs []Location) Location {
	return &OrderLocation{Locations: locs}
}

// Map the sequence at the pointing location.
func (loc OrderLocation) Map(seq Sequence) Sequence {
	r := make([]byte, loc.Len())
	i := 0
	for _, l := range loc.Locations {
		copy(r[i:], l.Map(seq).Bytes())
		i += l.Len()
	}
	return Seq(r)
}

// Len returns the length spanned by the location.
func (loc OrderLocation) Len() int {
	length := 0
	for _, l := range loc.Locations {
		length += l.Len()
	}
	return length
}

// Format the location.
func (loc OrderLocation) Format() string {
	tmp := make([]string, len(loc.Locations))
	for i := range loc.Locations {
		tmp[i] = loc.Locations[i].Format()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

// Shift the location position[s] if needed.
func (loc *OrderLocation) Shift(pos, n int) {
	for i := range loc.Locations {
		loc.Locations[i].Shift(pos, n)
	}
}

// Convert the given local index to a global index.
func (loc OrderLocation) Convert(index int) int {
	index = fixIndex(index, loc.Len())
	for _, l := range loc.Locations {
		if index < l.Len() {
			return l.Convert(index)
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
	locs := make([]Location, len(children))
	for i, child := range children {
		locs[i] = child.Value.(Location)
	}
	loc := NewJoinLocation(locs)
	result.SetValue(loc)
	return nil
})

// OrderLocationParser attempts to parse a OrderLocation.
var OrderLocationParser = pars.Seq(
	"order(", pars.Delim(&LocationParser, locationDelimiter), ')',
).Map(func(result *pars.Result) error {
	children := result.Children[1].Children
	locs := make([]Location, len(children))
	for i, child := range children {
		locs[i] = child.Value.(Location)
	}
	loc := NewOrderLocation(locs)
	result.SetValue(loc)
	return nil
})

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
