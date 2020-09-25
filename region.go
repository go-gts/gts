package gts

import (
	"fmt"

	"github.com/go-gts/gts/utils"
	"github.com/go-pars/pars"
)

// Modifier is an interface required to modify coordinate regions.
type Modifier interface {
	Apply(head, tail int) (int, int)
	Complement() Modifier
	fmt.Stringer
}

// Head collapses a region onto its head, offset by the value given.
type Head int

// Apply the modifier to the given bounds.
func (mod Head) Apply(head, tail int) (int, int) {
	n := int(mod) + head
	return n, n
}

// Complement returns the equivalent Modifier for the complement strand.
func (mod Head) Complement() Modifier {
	return Tail(-mod)
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
	n := int(mod) + tail
	return n, n
}

// Complement returns the equivalent Modifier for the complement strand.
func (mod Tail) Complement() Modifier {
	return Head(-mod)
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
	p, q := utils.Unpack(mod)
	return head + p, tail + q
}

// Complement returns the equivalent Modifier for the complement strand.
func (mod HeadTail) Complement() Modifier {
	p, q := utils.Unpack(mod)
	return HeadTail{-q, -p}
}

// String returns the textual representation of the Modifier.
func (mod HeadTail) String() string {
	p, q := utils.Unpack(mod)
	return fmt.Sprintf("%s..%s", Head(p), Tail(q))
}

// HeadHead offsets the head coordinate by the values given.
type HeadHead [2]int

// Apply the modifier to the given bounds.
func (mod HeadHead) Apply(head, tail int) (int, int) {
	p, q := utils.Unpack(mod)
	return head + p, head + q
}

// Complement returns the equivalent Modifier for the complement strand.
func (mod HeadHead) Complement() Modifier {
	p, q := utils.Unpack(mod)
	return TailTail{-q, -p}
}

// String returns the textual representation of the Modifier.
func (mod HeadHead) String() string {
	p, q := utils.Unpack(mod)
	return fmt.Sprintf("%s..%s", Head(p), Head(q))
}

// TailTail offsets the tail coordinate by the values given.
type TailTail [2]int

// Apply the modifier to the given bounds.
func (mod TailTail) Apply(head, tail int) (int, int) {
	p, q := utils.Unpack(mod)
	return tail + p, tail + q
}

// Complement returns the equivalent Modifier for the complement strand.
func (mod TailTail) Complement() Modifier {
	p, q := utils.Unpack(mod)
	return HeadHead{-q, -p}
}

// String returns the textual representation of the Modifier.
func (mod TailTail) String() string {
	p, q := utils.Unpack(mod)
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

// Region represents a coordinate region which can be resized and used to
// locate the subsequence within a given sequence corresponding to the region
// that is being represented.
type Region interface {
	Len() int
	Resize(mod Modifier) Region
	Complement() Region
	Locate(seq Sequence) Sequence
}

// Forward represents a Region on the forward strand.
type Forward [2]int

// Len returns the length spanned by the Region.
func (r Forward) Len() int {
	head, tail := utils.Unpack(r)
	return tail - head
}

// Resize the region using the given Modifier.
func (r Forward) Resize(mod Modifier) Region {
	head, tail := utils.Unpack(r)
	head, tail = mod.Apply(head, tail)
	if head < tail {
		return Forward{head, tail}
	}
	return Forward{head, head}
}

// Complement returns the equivalent Region on the complement strand.
func (r Forward) Complement() Region {
	return Backward(r)
}

// Locate the subsequence corresponding to the region in the given sequence.
func (r Forward) Locate(seq Sequence) Sequence {
	head, tail := utils.Unpack(r)
	return Slice(seq, head, tail)
}

// Backward represents a Region on the backward strand.
type Backward [2]int

// Len returns the length spanned by the Region.
func (r Backward) Len() int {
	head, tail := utils.Unpack(r)
	return tail - head
}

// Resize the region using the given Modifier.
func (r Backward) Resize(mod Modifier) Region {
	head, tail := utils.Unpack(r)
	head, tail = mod.Complement().Apply(head, tail)
	if head < tail {
		return Backward{head, tail}
	}
	return Backward{tail, tail}
}

// Complement returns the equivalent Region on the complement strand.
func (r Backward) Complement() Region {
	return Forward(r)
}

// Locate the subsequence corresponding to the region in the given sequence.
func (r Backward) Locate(seq Sequence) Sequence {
	head, tail := utils.Unpack(r)
	return Reverse(Complement(Slice(seq, head, tail)))
}

// Regions is a slice or Region objects.
type Regions []Region

// Len returns the length spanned by the Region.
func (rr Regions) Len() int {
	total := 0
	for _, r := range rr {
		total += r.Len()
	}
	return total
}

// Resize the region using the given Modifier.
func (rr Regions) Resize(mod Modifier) Region {
	ret := make(Regions, len(rr))
	copy(ret, rr)

	// Compute the lower and upper bounds from the head position.
	// This will unify the actual resize logic.
	lower, upper := 0, 0
	switch mod := mod.(type) {
	case Head:
		lower = int(mod)
		upper = lower
	case Tail:
		lower = int(mod) + ret.Len()
		upper = lower
	case HeadHead:
		lower, upper = utils.Unpack(mod)
	case HeadTail:
		lower, upper = utils.Unpack(mod)
		upper += ret.Len()
	case TailTail:
		lower, upper = utils.Unpack(mod)
		lower += ret.Len()
		upper += ret.Len()
	}

	left, right := 0, 0
	for k := 0; k+1 < len(rr); k++ {
		n := rr[k].Len()
		if n < lower {
			left = k + 1
			lower -= n
		}
		if n < upper {
			right = k + 1
			upper -= n
		}
	}

	switch utils.Compare(left, right) {
	case 1:
		return ret[left].Resize(Head(lower))
	case 0:
		return ret[left].Resize(HeadHead{lower, upper})
	default:
		ret[left] = ret[left].Resize(HeadTail{lower, 0})
		ret[right] = ret[right].Resize(HeadHead{0, upper})
		return ret[left : right+1]
	}
}

// Complement returns the equivalent Region on the complement strand.
func (rr Regions) Complement() Region {
	ret := make(Regions, len(rr))
	for i, r := range rr {
		// Flip the order of regions.
		ret[len(rr)-i-1] = r.Complement()
	}
	return ret
}

// Locate the subsequence corresponding to the region in the given sequence.
func (rr Regions) Locate(seq Sequence) Sequence {
	seqs := make([]Sequence, len(rr))
	for i, r := range rr {
		seqs[i] = r.Locate(seq)
	}
	return Concat(seqs...)
}
