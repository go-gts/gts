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

type Locator interface {
	Locate(string) string
	Format() string
	Shift(Shifter)
}

var locatableParser pars.Parser

type PointLocator struct {
	Position int
}

func NewPointLocator(pos int) Locator {
	return &PointLocator{Position: pos}
}

func (location PointLocator) Locate(s string) string {
	return s[location.Position-1 : location.Position]
}

func (location PointLocator) Format() string {
	return strconv.Itoa(location.Position)
}

func (location *PointLocator) Shift(shifter Shifter) {
	if shifter.Position <= location.Position {
		location.Position += shifter.Amount
	}
}

var pointLocatorParser = pars.Integer.Map(func(result *pars.Result) error {
	n, err := strconv.Atoi(result.Value.(string))
	if err != nil {
		return err
	}
	result.Value = n
	result.Children = nil
	return nil
})

type RangeLocator struct {
	Start    int
	End      int
	Partial5 bool
	Partial3 bool
}

func NewRangeLocator(start, end int, partial ...bool) Locator {
	return &RangeLocator{Start: start, End: end}
}

func NewPartialRangeLocator(start, end int, p5, p3 bool) Locator {
	return &RangeLocator{
		Start:    start,
		End:      end,
		Partial5: p5,
		Partial3: p3,
	}
}

func (location RangeLocator) Locate(s string) string {
	return s[location.Start-1 : location.End]
}

func (location RangeLocator) Format() string {
	p5, p3 := "", ""
	if location.Partial5 {
		p5 = "<"
	}
	if location.Partial3 {
		p3 = ">"
	}
	return fmt.Sprintf("%s%d..%s%d", p5, location.Start, p3, location.End)
}

func (location *RangeLocator) Shift(shifter Shifter) {
	if shifter.Position <= location.Start {
		location.Start += shifter.Amount
	}
	if shifter.Position <= location.End {
		location.End += shifter.Amount
	}
}

var rangeLocatorParser = pars.Seq(
	pars.Try('<'),
	pars.Integer.Map(pars.Atoi),
	"..",
	pars.Try('>'),
	pars.Integer.Map(pars.Atoi),
	pars.Try('>'), // Possibly required for some legacy entries.
).Map(func(result *pars.Result) error {
	result.Value = NewPartialRangeLocator(
		result.Children[1].Value.(int),
		result.Children[4].Value.(int),
		result.Children[0].Value != nil,
		result.Children[3].Value != nil || result.Children[5].Value != nil,
	)
	result.Children = nil
	return nil
})

type AmbiguousLocator struct {
	Start int
	End   int
}

func NewAmbiguousLocator(start, end int) Locator {
	return &AmbiguousLocator{Start: start, End: end}
}

func (location AmbiguousLocator) Locate(s string) string {
	return s[location.Start-1 : location.End]
}

func (location AmbiguousLocator) Format() string {
	return fmt.Sprintf("%d.%d", location.Start, location.End)
}

func (location *AmbiguousLocator) Shift(shifter Shifter) {
	if shifter.Position <= location.Start {
		location.Start += shifter.Amount
	}
	if shifter.Position <= location.End {
		location.End += shifter.Amount
	}
}

var ambiguousLocatorParser = pars.Seq(
	pars.Integer.Map(pars.Atoi), '.', pars.Integer.Map(pars.Atoi),
).Map(func(result *pars.Result) error {
	result.Value = NewAmbiguousLocator(
		result.Children[0].Value.(int),
		result.Children[2].Value.(int),
	)
	result.Children = nil
	return nil
})

type BetweenLocator struct {
	Start int
	End   int
}

func NewBetweenLocator(start, end int) Locator {
	return &BetweenLocator{Start: start, End: end}
}

func (location BetweenLocator) Locate(s string) string {
	return ""
}

func (location BetweenLocator) Format() string {
	return fmt.Sprintf("%d^%d", location.Start, location.End)
}

func (location *BetweenLocator) Shift(shifter Shifter) {
	if shifter.Position <= location.Start {
		location.Start += shifter.Amount
	}
	if shifter.Position <= location.End {
		location.End += shifter.Amount
	}
}

var betweenLocatorParser = pars.Seq(
	pars.Integer.Map(pars.Atoi), '^', pars.Integer.Map(pars.Atoi),
).Map(func(result *pars.Result) error {
	result.Value = NewBetweenLocator(
		result.Children[0].Value.(int),
		result.Children[2].Value.(int),
	)
	result.Children = nil
	return nil
})

type ComplementLocator struct {
	Locator Locator
}

func NewComplementLocator(location Locator) Locator {
	return &ComplementLocator{Locator: location}
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
	case 'A':
		return 'T'
	case 'T':
		return 'A'
	case 'G':
		return 'C'
	case 'C':
		return 'G'
	default:
		return b
	}
}

func (location ComplementLocator) Locate(s string) string {
	return strings.Map(complementBase, location.Locator.Locate(s))
}

func (location ComplementLocator) Format() string {
	return fmt.Sprintf("complement(%s)", location.Locator.Format())
}

func (location *ComplementLocator) Shift(shifter Shifter) {
	location.Locator.Shift(shifter)
}

var complementLocatorParser = pars.Seq(
	"complement(", &locatableParser, ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	result.Value = NewComplementLocator(result.Value.(Locator))
	result.Children = nil
	return nil
})

type JoinLocator struct {
	Locators []Locator
}

func NewJoinLocator(locations []Locator) Locator {
	return &JoinLocator{Locators: locations}
}

func (location JoinLocator) Locate(s string) string {
	tmp := make([]string, len(location.Locators))
	for i := range location.Locators {
		tmp[i] = location.Locators[i].Locate(s)
	}
	return strings.Join(tmp, "")
}

func (location JoinLocator) Format() string {
	tmp := make([]string, len(location.Locators))
	for i := range location.Locators {
		tmp[i] = location.Locators[i].Format()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

func (location *JoinLocator) Shift(shifter Shifter) {
	for i := range location.Locators {
		location.Locators[i].Shift(shifter)
	}
}

var joinLocatorParser = pars.Seq(
	"join(", pars.Delim(&locatableParser, ','), ')',
).Map(func(result *pars.Result) error {
	children := result.Children[1].Children
	locations := make([]Locator, len(children))
	for i, child := range children {
		locations[i] = child.Value.(Locator)
	}
	result.Value = NewJoinLocator(locations)
	result.Children = nil
	return nil
})

type OrderLocator struct {
	Locators []Locator
}

func NewOrderLocator(locations []Locator) Locator {
	return &OrderLocator{Locators: locations}
}

func (location OrderLocator) Locate(s string) string {
	tmp := make([]string, len(location.Locators))
	for i := range location.Locators {
		tmp[i] = location.Locators[i].Locate(s)
	}
	return strings.Join(tmp, "")
}

func (location OrderLocator) Format() string {
	tmp := make([]string, len(location.Locators))
	for i := range location.Locators {
		tmp[i] = location.Locators[i].Format()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

func (location *OrderLocator) Shift(shifter Shifter) {
	for i := range location.Locators {
		location.Locators[i].Shift(shifter)
	}
}

var orderLocatorParser = pars.Seq(
	"order(", pars.Delim(&locatableParser, ','), ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	locations := make([]Locator, len(result.Children))
	for i, child := range result.Children {
		locations[i] = child.Value.(Locator)
	}
	result.Value = NewOrderLocator(locations)
	result.Children = nil
	return nil
})

func init() {
	locatableParser = pars.Any(
		rangeLocatorParser,
		orderLocatorParser,
		joinLocatorParser,
		complementLocatorParser,
		ambiguousLocatorParser,
		betweenLocatorParser,
		pointLocatorParser,
	)
}
