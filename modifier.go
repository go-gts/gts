package gts

import (
	"fmt"

	"github.com/go-pars/pars"
)

// Modifier is an interface required to modify coordinate regions.
type Modifier interface {
	Apply(head, tail int) (int, int)
	fmt.Stringer
}

// Head collapses a region onto its head, offset by the value given.
type Head int

// Apply the modifier to the given bounds.
func (mod Head) Apply(head, tail int) (int, int) {
	p := int(mod)
	if tail < head {
		head, tail = mod.Apply(-head, -tail)
		return -head, -tail
	}
	head += p
	return head, head
}

// String returns the textual representation of the Modifier.
func (mod Head) String() string {
	if mod == 0 {
		return "^"
	}
	return fmt.Sprintf("^%+d", mod)
}

// Tail collapses a region onto its tail, offset by the value given.
type Tail int

// Apply the modifier to the given bounds.
func (mod Tail) Apply(head, tail int) (int, int) {
	q := int(mod)
	if tail < head {
		head, tail = mod.Apply(-head, -tail)
		return -head, -tail
	}
	tail += q
	return tail, tail
}

// String returns the textual representation of the Modifier.
func (mod Tail) String() string {
	if mod == 0 {
		return "$"
	}
	return fmt.Sprintf("$%+d", mod)
}

// HeadTail offsets the head and tail coordinates by the values given.
type HeadTail [2]int

// Apply the modifier to the given bounds.
func (mod HeadTail) Apply(head, tail int) (int, int) {
	p, q := Unpack(mod)

	// Direction is backward. Assume complement.
	if tail < head {
		head, tail = mod.Apply(-head, -tail)
		return -head, -tail
	}

	head += p
	tail += q
	return head, Max(head, tail)
}

// String returns the textual representation of the Modifier.
func (mod HeadTail) String() string {
	p, q := Unpack(mod)
	return fmt.Sprintf("%s..%s", Head(p), Tail(q))
}

// HeadHead offsets the head coordinate by the values given.
type HeadHead [2]int

// Apply the modifier to the given bounds.
func (mod HeadHead) Apply(head, tail int) (int, int) {
	p, q := Unpack(mod)

	if tail < head {
		head, tail = mod.Apply(-head, -tail)
		return -head, -tail
	}

	tail = head + q
	head += p
	return head, Max(head, tail)

}

// String returns the textual representation of the Modifier.
func (mod HeadHead) String() string {
	p, q := Unpack(mod)
	return fmt.Sprintf("%s..%s", Head(p), Head(q))
}

// TailTail offsets the head coordinate by the values given.
type TailTail [2]int

// Apply the modifier to the given bounds.
func (mod TailTail) Apply(head, tail int) (int, int) {
	p, q := Unpack(mod)

	if tail < head {
		head, tail = mod.Apply(-head, -tail)
		return -head, -tail
	}

	head = tail + p
	tail += q
	return head, Max(head, tail)
}

// String returns the textual representation of the Modifier.
func (mod TailTail) String() string {
	p, q := Unpack(mod)
	return fmt.Sprintf("%s..%s", Tail(p), Tail(q))
}

var parseHead = pars.Any(
	pars.Seq('^', pars.Int).Child(1),
	pars.Byte('^').Bind(0),
).Map(func(result *pars.Result) error {
	n := result.Value.(int)
	result.SetValue(Head(n))
	return nil
})

var parseTail = pars.Any(
	pars.Seq('$', pars.Int).Child(1),
	pars.Byte('$').Bind(0),
).Map(func(result *pars.Result) error {
	n := result.Value.(int)
	result.SetValue(Tail(n))
	return nil
})

func mapHeadTail(result *pars.Result) error {
	p := int(result.Children[0].Value.(Head))
	q := int(result.Children[2].Value.(Tail))
	result.SetValue(HeadTail{p, q})
	return nil
}

var parseHeadTail = pars.Seq(parseHead, "..", parseTail).Map(mapHeadTail)

func mapHeadHead(result *pars.Result) error {
	p := int(result.Children[0].Value.(Head))
	q := int(result.Children[2].Value.(Head))
	result.SetValue(HeadHead{p, q})
	return nil
}

var parseHeadHead = pars.Seq(parseHead, "..", parseHead).Map(mapHeadHead)

func mapTailTail(result *pars.Result) error {
	p := int(result.Children[0].Value.(Tail))
	q := int(result.Children[2].Value.(Tail))
	result.SetValue(TailTail{p, q})
	return nil
}

var parseTailTail = pars.Seq(parseTail, "..", parseTail).Map(mapTailTail)

var parseModifier = pars.Any(
	parseHeadTail,
	parseHeadHead,
	parseTailTail,
	parseHead,
	parseTail,
)

// AsModifier interprets the given string as a Modifier.
func AsModifier(s string) (Modifier, error) {
	result, err := pars.Exact(parseModifier).Parse(pars.FromString(s))
	if err != nil {
		return nil, err
	}
	return result.Value.(Modifier), nil
}
