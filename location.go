package gt1

import (
	"errors"
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

	// String satisfies the fmt.Stringer interface.
	String() string

	// Shift the location by the given amount if needed.
	// Returns false if the shift invalidates the location.
	Shift(offset, amount int) bool

	// Map the given local index to a global index.
	Map(index int) int
}

// LocationLess tests if the one location is smaller than the other.
func LocationLess(a, b Location) bool {
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

// Locate the sequence at the pointing location.
func (loc PointLocation) Locate(seq Sequence) Sequence {
	return Slice(seq, loc.Position, loc.Position+1)
}

// Len returns the length spanned by the location.
func (loc PointLocation) Len() int {
	return 1
}

// String satisfies the fmt.Stringer interface.
func (loc PointLocation) String() string {
	return strconv.Itoa(loc.Position + 1)
}

// Shift the location position[s] if needed.
// Returns false if the shift invalidates the location.
func (loc *PointLocation) Shift(offset, amount int) bool {
	if amount == 0 || loc.Position < offset {
		return true
	}
	if amount < 0 && loc.Position < offset-amount {
		return false
	}
	loc.Position += amount
	return true
}

// Map the given local index to a global index.
func (loc PointLocation) Map(index int) int {
	index = fixIndex(index, loc.Len())
	return loc.Position + index
}

func shiftRange(a, b, i, n int) (int, int, bool) {
	switch {
	case n > 0:
		if i <= a {
			a += n
		}
		if i <= b {
			b += n
		}
		return a, b, true
	case n < 0:
		c, d := a, b
		if i-n <= c {
			c += n
		}
		if i-n <= d {
			d += n
		}
		if c < d-1 {
			return c, d, true
		}
		return a, b, false
	default:
		return a, b, true
	}
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
func (loc RangeLocation) Locate(seq Sequence) Sequence {
	return Slice(seq, loc.Start, loc.End)
}

// Len returns the length spanned by the location.
func (loc RangeLocation) Len() int {
	return loc.End - loc.Start
}

// String satisfies the fmt.Stringer interface.
func (loc RangeLocation) String() string {
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
// Returns false if the shift invalidates the location.
func (loc *RangeLocation) Shift(offset, amount int) (ok bool) {
	loc.Start, loc.End, ok = shiftRange(loc.Start, loc.End, offset, amount)
	return
}

// Map the given local index to a global index.
func (loc RangeLocation) Map(index int) int {
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

// Locate the sequence at the pointing location.
func (loc AmbiguousLocation) Locate(seq Sequence) Sequence {
	return Slice(seq, loc.Start, loc.End)
}

// Len returns the length spanned by the location.
func (loc AmbiguousLocation) Len() int {
	return loc.End - loc.Start
}

// String satisfies the fmt.Stringer interface.
func (loc AmbiguousLocation) String() string {
	return fmt.Sprintf("%d.%d", loc.Start+1, loc.End)
}

// Shift the location position[s] if needed.
// Returns false if the shift invalidates the location.
func (loc *AmbiguousLocation) Shift(offset, amount int) (ok bool) {
	loc.Start, loc.End, ok = shiftRange(loc.Start, loc.End, offset, amount)
	return
}

// Map the given local index to a global index.
func (loc AmbiguousLocation) Map(index int) int {
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

// Locate the sequence at the pointing location.
func (loc BetweenLocation) Locate(seq Sequence) Sequence {
	return Slice(seq, loc.Start, loc.End)
}

// Len returns the length spanned by the location.
func (loc BetweenLocation) Len() int {
	return loc.End - loc.Start
}

// String satisfies the fmt.Stringer interface.
func (loc BetweenLocation) String() string {
	return fmt.Sprintf("%d^%d", loc.Start+1, loc.End)
}

// Shift the location position[s] if needed.
// Returns false if the shift invalidates the location.
func (loc *BetweenLocation) Shift(offset, amount int) (ok bool) {
	loc.Start, loc.End, ok = shiftRange(loc.Start, loc.End, offset, amount)
	return
}

// Map the given local index to a global index.
func (loc BetweenLocation) Map(index int) int {
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

// Locate the sequence at the pointing location.
func (loc ComplementLocation) Locate(seq Sequence) Sequence {
	return Complement(loc.Location.Locate(seq))
}

// Len returns the length spanned by the location.
func (loc ComplementLocation) Len() int {
	return loc.Location.Len()
}

// String satisfies the fmt.Stringer interface.
func (loc ComplementLocation) String() string {
	return fmt.Sprintf("complement(%s)", loc.Location.String())
}

// Shift the location position[s] if needed.
// Returns false if the shift invalidates the location.
func (loc *ComplementLocation) Shift(offset, amount int) bool {
	return loc.Location.Shift(offset, amount)
}

// Map the given local index to a global index.
func (loc ComplementLocation) Map(index int) int {
	return loc.Location.Map(index)
}

// JoinLocation represents multiple joined locations.
type JoinLocation struct {
	Locations []Location
}

// NewJoinLocation creates a new JoinLocation.
func NewJoinLocation(locs []Location) Location {
	return &JoinLocation{Locations: locs}
}

// Locate the sequence at the pointing location.
func (loc JoinLocation) Locate(seq Sequence) Sequence {
	r := make([]byte, loc.Len())
	i := 0
	for _, l := range loc.Locations {
		copy(r[i:], l.Locate(seq).Bytes())
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

// String satisfies the fmt.Stringer interface.
func (loc JoinLocation) String() string {
	tmp := make([]string, len(loc.Locations))
	for i := range loc.Locations {
		tmp[i] = loc.Locations[i].String()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

// Shift the location position[s] if needed.
// Returns false if the shift invalidates the location.
func (loc *JoinLocation) Shift(pos, n int) bool {
	ok := true
	for i := range loc.Locations {
		if !loc.Locations[i].Shift(pos, n) {
			ok = false
		}
	}
	return ok
}

// Map the given local index to a global index.
func (loc JoinLocation) Map(index int) int {
	index = fixIndex(index, loc.Len())
	for _, l := range loc.Locations {
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
func NewOrderLocation(locs []Location) Location {
	return &OrderLocation{Locations: locs}
}

// Locate the sequence at the pointing location.
func (loc OrderLocation) Locate(seq Sequence) Sequence {
	r := make([]byte, loc.Len())
	i := 0
	for _, l := range loc.Locations {
		copy(r[i:], l.Locate(seq).Bytes())
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

// String satisfies the fmt.Stringer interface.
func (loc OrderLocation) String() string {
	tmp := make([]string, len(loc.Locations))
	for i := range loc.Locations {
		tmp[i] = loc.Locations[i].String()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

// Shift the location position[s] if needed.
// Returns false if the shift invalidates the location.
func (loc *OrderLocation) Shift(pos, n int) bool {
	ok := true
	for i := range loc.Locations {
		if !loc.Locations[i].Shift(pos, n) {
			ok = false
		}
	}
	return ok
}

// Map the given local index to a global index.
func (loc OrderLocation) Map(index int) int {
	index = fixIndex(index, loc.Len())
	for _, l := range loc.Locations {
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

var errNotLocation = errors.New("string is not a Location")

// AsLocation will attempt to interpret the given string as a Location.
func AsLocation(s string) (Location, error) {
	state := pars.FromString(s)
	result := pars.Result{}
	parser := pars.Exact(LocationParser).Error(errNotLocation)
	if err := parser(state, &result); err != nil {
		return nil, err
	}
	return result.Value.(Location), nil
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
