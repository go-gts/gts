package gts

import (
	"sort"
)

// Region represents a coordinate region which can be resized and used to
// locate the subsequence within a given sequence corresponding to the region
// that is being represented.
type Region interface {
	Len() int
	Head() int
	Tail() int
	Resize(mod Modifier) Region
	Complement() Region
	Within(lower, upper int) bool
	Overlap(lower, upper int) bool
	Locate(seq Sequence) Sequence
}

// Segment represents a contiguous region.
type Segment [2]int

// Len returns the length spanned by the region.
func (s Segment) Len() int {
	start, end := Unpack(s)
	return Abs(end - start)
}

// Head returns the 5' boundary of the region.
func (s Segment) Head() int {
	return s[0]
}

// Tail returns the 3' boundary of the region.
func (s Segment) Tail() int {
	return s[1]
}

// Resize the region using the given Modifier.
func (s Segment) Resize(mod Modifier) Region {
	head, tail := Unpack(s)
	head, tail = mod.Apply(head, tail)
	return Segment{head, tail}
}

// Complement returns the equivalent region for the complement strand.
func (s Segment) Complement() Region {
	head, tail := Unpack(s)
	return Segment{tail, head}
}

// Within checks if the region is within the bounds of the given segment.
func (s Segment) Within(lower, upper int) bool {
	head, tail := Unpack(s)
	if tail < head {
		head, tail = tail, head
	}
	return sort.IntsAreSorted([]int{lower, head, tail, upper})
}

// Overlap checks if the region overlaps with the bounds of the given segment.
func (s Segment) Overlap(lower, upper int) bool {
	head, tail := Unpack(s)
	if tail < head {
		head, tail = tail, head
	}
	return (head < upper) && (lower < tail)
}

// Locate the subsequence corresponding to the region in the given sequence.
func (s Segment) Locate(seq Sequence) Sequence {
	head, tail := Unpack(s)
	if tail < head {
		return Reverse(Complement(Slice(seq, tail, head)))
	}
	return Slice(seq, head, tail)
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

// Head returns the 5' boundary of the region.
func (rr Regions) Head() int {
	if len(rr) > 0 {
		return rr[0].Head()
	}
	return 0
}

// Tail returns the 3' boundary of the region.
func (rr Regions) Tail() int {
	if len(rr) > 0 {
		return rr[len(rr)-1].Tail()
	}
	return 0
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
		lower, upper = Unpack(mod)
	case HeadTail:
		lower, upper = Unpack(mod)
		upper += ret.Len()
	case TailTail:
		lower, upper = Unpack(mod)
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

	switch Compare(left, right) {
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

// Within checks if the region is within the bounds of the given segment.
func (rr Regions) Within(lower, upper int) bool {
	for _, r := range rr {
		if !r.Within(lower, upper) {
			return false
		}
	}
	return true
}

// Overlap checks if the region overlaps with the bounds of the given segment.
func (rr Regions) Overlap(lower, upper int) bool {
	for _, r := range rr {
		if r.Overlap(lower, upper) {
			return true
		}
	}
	return false
}

// Locate the subsequence corresponding to the region in the given sequence.
func (rr Regions) Locate(seq Sequence) Sequence {
	seqs := make([]Sequence, len(rr))
	for i, r := range rr {
		seqs[i] = r.Locate(seq)
	}
	return Concat(seqs...)
}

// BySegment attaches the methods of sort.Interface to []Segment, sorting in
// increasing order.
type BySegment []Segment

// Len is the number of elements in the collection.
func (ss BySegment) Len() int {
	return len(ss)
}

// Less reports whether the element with index i should sort before the element
// with index j.
func (ss BySegment) Less(i, j int) bool {
	l, r := ss[i], ss[j]
	if l[1] < l[0] {
		l[0], l[1] = l[1], l[0]
	}
	if r[1] < r[0] {
		r[0], r[1] = r[1], r[0]
	}
	if l[0] < r[0] {
		return true
	}
	if r[0] < l[0] {
		return false
	}
	if l[1] < r[1] {
		return true
	}
	return false
}

// Swap the elements with indexes i and j.
func (ss BySegment) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

func flattenRegion(arg Region) []Segment {
	switch rr := arg.(type) {
	case Regions:
		ss := []Segment{}
		for _, r := range rr {
			ss = append(ss, flattenRegion(r)...)
		}
		return ss
	default:
		s := rr.(Segment)
		if s[1] < s[0] {
			s = Segment{s[1], s[0]}
		}
		return []Segment{s}
	}
}

// Minimize the representation of the given region. A minimized region will
// be flattened, sorted, and overlapping areas removed.
func Minimize(arg Region) []Segment {
	ss := flattenRegion(arg)
	sort.Sort(BySegment(ss))
	i := 0
	for i < len(ss)-1 {
		l, r := ss[i], ss[i+1]
		if l[1] < r[0] {
			i++
		} else {
			ss[i] = Segment{Min(l[0], r[0]), Max(l[1], r[1])}
			copy(ss[i+1:], ss[i+2:])
			ss = ss[:len(ss)-1]
		}
	}
	return ss
}
