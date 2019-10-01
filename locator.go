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
	Locate([]byte) []byte
	Length() int
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

func (locator PointLocator) Locate(s []byte) []byte {
	return s[locator.Position : locator.Position+1]
}

func (locator PointLocator) Length() int {
	return 1
}

func (locator PointLocator) Format() string {
	return strconv.Itoa(locator.Position + 1)
}

func (locator *PointLocator) Shift(shifter Shifter) {
	if shifter.Position <= locator.Position {
		locator.Position += shifter.Amount
	}
}

var pointLocatorParser = pars.Integer.Map(func(result *pars.Result) error {
	n, err := strconv.Atoi(result.Value.(string))
	if err != nil {
		return err
	}
	result.Value = NewPointLocator(n - 1)
	result.Children = nil
	return nil
})

type RangeLocator struct {
	Start    int
	End      int
	Partial5 bool
	Partial3 bool
}

func NewRangeLocator(start, end int) Locator {
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

func (locator RangeLocator) Locate(s []byte) []byte {
	return s[locator.Start:locator.End]
}

func (locator RangeLocator) Length() int {
	return locator.End - locator.Start
}

func (locator RangeLocator) Format() string {
	p5, p3 := "", ""
	if locator.Partial5 {
		p5 = "<"
	}
	if locator.Partial3 {
		p3 = ">"
	}
	return fmt.Sprintf("%s%d..%s%d", p5, locator.Start+1, p3, locator.End)
}

func (locator *RangeLocator) Shift(shifter Shifter) {
	if shifter.Position <= locator.Start {
		locator.Start += shifter.Amount
	}
	if shifter.Position <= locator.End {
		locator.End += shifter.Amount
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
		result.Children[1].Value.(int)-1,
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

func (locator AmbiguousLocator) Locate(s []byte) []byte {
	return s[locator.Start:locator.End]
}

func (locator AmbiguousLocator) Length() int {
	return locator.End - locator.Start
}

func (locator AmbiguousLocator) Format() string {
	return fmt.Sprintf("%d.%d", locator.Start+1, locator.End)
}

func (locator *AmbiguousLocator) Shift(shifter Shifter) {
	if shifter.Position <= locator.Start {
		locator.Start += shifter.Amount
	}
	if shifter.Position <= locator.End {
		locator.End += shifter.Amount
	}
}

var ambiguousLocatorParser = pars.Seq(
	pars.Integer.Map(pars.Atoi), '.', pars.Integer.Map(pars.Atoi),
).Map(func(result *pars.Result) error {
	result.Value = NewAmbiguousLocator(
		result.Children[0].Value.(int)-1,
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

func (locator BetweenLocator) Locate(s []byte) []byte {
	return s[locator.Start:locator.End]
}

func (locator BetweenLocator) Length() int {
	return locator.End - locator.Start
}

func (locator BetweenLocator) Format() string {
	return fmt.Sprintf("%d^%d", locator.Start+1, locator.End)
}

func (locator *BetweenLocator) Shift(shifter Shifter) {
	if shifter.Position <= locator.Start {
		locator.Start += shifter.Amount
	}
	if shifter.Position <= locator.End {
		locator.End += shifter.Amount
	}
}

var betweenLocatorParser = pars.Seq(
	pars.Integer.Map(pars.Atoi), '^', pars.Integer.Map(pars.Atoi),
).Map(func(result *pars.Result) error {
	result.Value = NewBetweenLocator(
		result.Children[0].Value.(int)-1,
		result.Children[2].Value.(int),
	)
	result.Children = nil
	return nil
})

type ComplementLocator struct {
	Locator Locator
}

func NewComplementLocator(locator Locator) Locator {
	return &ComplementLocator{Locator: locator}
}

func complementBase(b byte) byte {
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

func (locator ComplementLocator) Locate(s []byte) []byte {
	r := make([]byte, len(s))
	for i, b := range s {
		r[i] = complementBase(b)
	}
	return r
}

func (locator ComplementLocator) Length() int {
	return locator.Locator.Length()
}

func (locator ComplementLocator) Format() string {
	return fmt.Sprintf("complement(%s)", locator.Locator.Format())
}

func (locator *ComplementLocator) Shift(shifter Shifter) {
	locator.Locator.Shift(shifter)
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

func NewJoinLocator(locators []Locator) Locator {
	return &JoinLocator{Locators: locators}
}

func (locator JoinLocator) Locate(s []byte) []byte {
	r := make([]byte, 0, locator.Length())
	for _, l := range locator.Locators {
		r = append(r, l.Locate(s)...)
	}
	return r
}

func (locator JoinLocator) Length() int {
	length := 0
	for _, l := range locator.Locators {
		length += l.Length()
	}
	return length
}

func (locator JoinLocator) Format() string {
	tmp := make([]string, len(locator.Locators))
	for i := range locator.Locators {
		tmp[i] = locator.Locators[i].Format()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

func (locator *JoinLocator) Shift(shifter Shifter) {
	for i := range locator.Locators {
		locator.Locators[i].Shift(shifter)
	}
}

var joinLocatorParser = pars.Seq(
	"join(", pars.Delim(&locatableParser, ','), ')',
).Map(func(result *pars.Result) error {
	children := result.Children[1].Children
	locators := make([]Locator, len(children))
	for i, child := range children {
		locators[i] = child.Value.(Locator)
	}
	result.Value = NewJoinLocator(locators)
	result.Children = nil
	return nil
})

type OrderLocator struct {
	Locators []Locator
}

func NewOrderLocator(locators []Locator) Locator {
	return &OrderLocator{Locators: locators}
}

func (locator OrderLocator) Locate(s []byte) []byte {
	r := make([]byte, 0, locator.Length())
	for _, l := range locator.Locators {
		r = append(r, l.Locate(s)...)
	}
	return r
}

func (locator OrderLocator) Length() int {
	length := 0
	for _, l := range locator.Locators {
		length += l.Length()
	}
	return length
}

func (locator OrderLocator) Format() string {
	tmp := make([]string, len(locator.Locators))
	for i := range locator.Locators {
		tmp[i] = locator.Locators[i].Format()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

func (locator *OrderLocator) Shift(shifter Shifter) {
	for i := range locator.Locators {
		locator.Locators[i].Shift(shifter)
	}
}

var orderLocatorParser = pars.Seq(
	"order(", pars.Delim(&locatableParser, ','), ')',
).Map(pars.Child(1)).Map(func(result *pars.Result) error {
	locators := make([]Locator, len(result.Children))
	for i, child := range result.Children {
		locators[i] = child.Value.(Locator)
	}
	result.Value = NewOrderLocator(locators)
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
