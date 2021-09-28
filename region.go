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
	Locate(seq Sequence) Sequence
	Crop(ff Features) Features
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

// Locate the subsequence corresponding to the region in the given sequence.
func (s Segment) Locate(seq Sequence) Sequence {
	head, tail := Unpack(s)
	if tail < head {
		return Apply(seq, Slicer(tail, head), Reverse, Complement)
	}
	return Slice(seq, head, tail)
}

// Crop the features to the region.
func (s Segment) Crop(ff Features) Features {
	head, tail := Unpack(s)
	if tail < head {
		gg := make(Features, len(ff))
		for i, f := range ff {
			gg[i] = Feature{f.Key, f.Loc.Complement(), f.Props.Clone()}
		}
		return Segment{tail, head}.Crop(gg)
	}
	ff = ff.Filter(Overlap(head, tail))
	gg := Features{}
	for _, f := range ff {
		loc := f.Loc.Expand(tail, intMin).Expand(0, -head)
		gg = gg.Insert(NewFeature(f.Key, loc, f.Props.Clone()))
	}
	return gg
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

// Locate the subsequence corresponding to the region in the given sequence.
func (rr Regions) Locate(seq Sequence) Sequence {
	seqs := make([]Sequence, len(rr))
	for i, r := range rr {
		seqs[i] = r.Locate(seq)
	}
	return Concat(seqs...)
}

// Crop the features to the region.
func (rr Regions) Crop(ff Features) Features {
	gg := Features{}
	ss := flattenRegion(rr)
	offset := 0
	for _, s := range ss {
		for _, f := range s.Crop(ff) {
			loc := f.Loc.Expand(0, offset)
			gg = gg.Insert(NewFeature(f.Key, loc, f.Props.Clone()))
		}
		offset += s.Len()
	}
	return gg
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

func invertSegments(ss []Segment, n int) []Segment {
	rr := make([]Segment, 0, len(ss)+1)
	start := 0
	for _, s := range ss {
		if start != s[0] {
			rr = append(rr, Segment{start, s[0]})
		}
		start = s[1]
	}
	if start != n {
		rr = append(rr, Segment{start, n})
	}
	return rr
}

func InvertLinear(r Region, n int) []Region {
	ss := Minimize(r)
	ss = invertSegments(ss, n)
	rr := make([]Region, len(ss))
	for i, s := range ss {
		rr[i] = s
	}
	return rr
}

func InvertCircular(r Region, n int) []Region {
	ss := Minimize(r)
	rr := InvertLinear(r, n)
	if ss[0][0] == 0 || ss[len(ss)-1][1] == n {
		return rr
	}
	rr[0] = Regions{rr[len(rr)-1], rr[0]}
	return rr[:len(rr)-1]
}
