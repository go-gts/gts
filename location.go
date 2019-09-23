package gd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ktnyt/pars"
)

type Shifter struct {
	Position int
	Amount   int
}

type Locatable interface {
	Locate(string) string
	Format() string
	Shift(Shifter)
}

var locatableParser pars.Parser

type PointLocation struct {
	Position int
}

func NewPointLocation(pos int) Locatable {
	return &PointLocation{Position: pos}
}

func (location PointLocation) Locate(s string) string {
	return s[location.Position : location.Position+1]
}

func (location PointLocation) Format() string {
	return strconv.Itoa(location.Position)
}

func (location *PointLocation) Shift(shifter Shifter) {
	if shifter.Position <= location.Position {
		location.Position += shifter.Amount
	}
}

var pointLocationParser = pars.Integer.Map(func(result *pars.Result) error {
	n, err := strconv.Atoi(result.Value.(string))
	if err != nil {
		return err
	}
	result.Value = n
	result.Children = nil
	return nil
})

type RangeLocation struct {
	Start    int
	End      int
	Partial5 bool
	Partial3 bool
}

func NewRangeLocation(start, end int, partial ...bool) Locatable {
	return &RangeLocation{Start: start, End: end}
}

func NewPartialRangeLocation(start, end int, p5, p3 bool) Locatable {
	return &RangeLocation{
		Start:    start,
		End:      end,
		Partial5: p5,
		Partial3: p3,
	}
}

func (location RangeLocation) Locate(s string) string {
	return s[location.Start:location.End]
}

func (location RangeLocation) Format() string {
	p5, p3 := "", ""
	if location.Partial5 {
		p5 = "<"
	}
	if location.Partial3 {
		p3 = ">"
	}
	return fmt.Sprintf("%s%d..%s%d", p5, location.Start, p3, location.End)
}

func (location *RangeLocation) Shift(shifter Shifter) {
	if shifter.Position <= location.Start {
		location.Start += shifter.Amount
	}
	if shifter.Position <= location.End {
		location.End += shifter.Amount
	}
}

var rangeLocationParser = pars.Seq(
	pars.Try('<'),
	pars.Integer.Map(pars.Atoi),
	"..",
	pars.Try('>'),
	pars.Integer.Map(pars.Atoi),
	pars.Try('>'), // Possibly required for some legacy entries.
).Map(func(result *pars.Result) error {
	result.Value = NewPartialRangeLocation(
		result.Children[1].Value.(int),
		result.Children[4].Value.(int),
		result.Children[0].Value != nil,
		result.Children[3].Value != nil || result.Children[5].Value != nil,
	)
	result.Children = nil
	return nil
})

type AmbiguousLocation struct {
	Start int
	End   int
}

func NewAmbiguousLocation(start, end int) Locatable {
	return &AmbiguousLocation{Start: start, End: end}
}

func (location AmbiguousLocation) Locate(s string) string {
	return s[location.Start:location.End]
}

func (location AmbiguousLocation) Format() string {
	return fmt.Sprintf("%d.%d", location.Start, location.End)
}

func (location *AmbiguousLocation) Shift(shifter Shifter) {
	if shifter.Position <= location.Start {
		location.Start += shifter.Amount
	}
	if shifter.Position <= location.End {
		location.End += shifter.Amount
	}
}

var ambiguousLocationParser = pars.Seq(
	pars.Integer.Map(pars.Atoi), '.', pars.Integer.Map(pars.Atoi),
).Map(func(result *pars.Result) error {
	result.Value = AmbiguousLocation{
		Start: result.Children[0].Value.(int),
		End:   result.Children[2].Value.(int),
	}
	result.Children = nil
	return nil
})

type ComplementLocation struct {
	Location Locatable
}

func NewComplementLocation(location Locatable) Locatable {
	return &ComplementLocation{Location: location}
}

func complementBase(b rune) rune {
	switch b {
	case 'a':
		return 't'
	case 't':
		return 'a'
	case 'g':
		return 'c'
	case 'c':
		return 'g'
	default:
		return b
	}
}

func (location ComplementLocation) Locate(s string) string {
	return strings.Map(complementBase, location.Location.Locate(s))
}

func (location ComplementLocation) Format() string {
	return fmt.Sprintf("complement(%s)", location.Location.Format())
}

func (location *ComplementLocation) Shift(shifter Shifter) {
	location.Location.Shift(shifter)
}

var complementLocationParser = pars.Seq(
	"complement(", &locatableParser, ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	result.Value = NewComplementLocation(result.Value.(Locatable))
	result.Children = nil
	return nil
})

type JoinLocation struct {
	Locations []Locatable
}

func NewJoinLocation(locations []Locatable) Locatable {
	return &JoinLocation{Locations: locations}
}

func (location JoinLocation) Locate(s string) string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Locate(s)
	}
	return strings.Join(tmp, "")
}

func (location JoinLocation) Format() string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Format()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

func (location *JoinLocation) Shift(shifter Shifter) {
	for i := range location.Locations {
		location.Locations[i].Shift(shifter)
	}
}

var joinLocationParser = pars.Seq(
	"join(", pars.Delim(&locatableParser, ','), ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	locations := make([]Locatable, len(result.Children))
	for i, child := range result.Children {
		locations[i] = child.Value.(Locatable)
	}
	result.Value = NewJoinLocation(locations)
	result.Children = nil
	return nil
})

type OrderLocation struct {
	Locations []Locatable
}

func NewOrderLocation(locations []Locatable) Locatable {
	return &OrderLocation{Locations: locations}
}

func (location OrderLocation) Locate(s string) string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Locate(s)
	}
	return strings.Join(tmp, "")
}

func (location OrderLocation) Format() string {
	tmp := make([]string, len(location.Locations))
	for i := range location.Locations {
		tmp[i] = location.Locations[i].Format()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

func (location *OrderLocation) Shift(shifter Shifter) {
	for i := range location.Locations {
		location.Locations[i].Shift(shifter)
	}
}

var orderLocationParser = pars.Seq(
	"order(", pars.Delim(&locatableParser, ','), ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	locations := make([]Locatable, len(result.Children))
	for i, child := range result.Children {
		locations[i] = child.Value.(Locatable)
	}
	result.Value = NewOrderLocation(locations)
	result.Children = nil
	return nil
})

func init() {
	locatableParser = pars.Any(
		orderLocationParser,
		joinLocationParser,
		complementLocationParser,
		rangeLocationParser,
		ambiguousLocationParser,
		pointLocationParser,
	)
}
