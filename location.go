package gts

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-ascii/ascii"
	"github.com/go-pars/pars"
)

func rangeCompare(s1, e1, s2, e2 int) int {
	if e1 < s1 {
		s1, e1 = e1, s1
	}
	if e2 < s2 {
		s2, e2 = e2, s2
	}
	switch {
	case s1 < s2:
		return -1
	case s2 < s1:
		return 1
	case e1 < e2:
		return -1
	case e2 < e1:
		return 1
	default:
		return 0
	}
}

func rangeWithin(s, e, l, u int) bool {
	if e < s {
		s, e = e, s
	}
	if u < l {
		l, u = u, l
	}
	return l <= s && e <= u
}

func rangeOverlap(s, e, l, u int) bool {
	if e < s {
		s, e = e, s
	}
	if u < l {
		l, u = u, l
	}
	return s < u && l < e
}

// LocationLess tests if location a is less than b.
func LocationLess(a, b Location) bool {
	if c, ok := a.(Complemented); ok {
		return LocationLess(c.Location, b)
	}

	if c, ok := b.(Complemented); ok {
		return LocationLess(a, c.Location)
	}

	if ll, ok := a.(locationSlice); ok {
		for _, l := range ll.slice() {
			if LocationLess(l, b) {
				return true
			}
		}
		return false
	}

	if ll, ok := b.(locationSlice); ok {
		for _, l := range ll.slice() {
			if !LocationLess(a, l) {
				return false
			}
		}
		return true
	}

	l, lok := a.(contiguousLocation)
	r, rok := b.(contiguousLocation)

	switch {
	case lok && rok:
		s1, e1 := l.span()
		s2, e2 := r.span()
		if cmp := rangeCompare(s1, e1, s2, e2); cmp != 0 {
			return cmp < 0
		}

		v, vok := a.(Ranged)
		u, uok := b.(Ranged)
		vp, up := 0, 0

		if vok {
			if v.Partial[0] {
				vp++
			}
			if v.Partial[1] {
				vp++
			}
		}

		if uok {
			if u.Partial[0] {
				up++
			}
			if u.Partial[1] {
				up++
			}
		}

		return vp < up

	case lok:
		return true

	default:
		return false
	}
}

// LocationWithin tests if the given location is within the given bounds.
func LocationWithin(loc Location, lower, upper int) bool {
	switch v := loc.(type) {
	case Complemented:
		return LocationWithin(v.Location, lower, upper)

	case locationSlice:
		for _, l := range v.slice() {
			if !LocationWithin(l, lower, upper) {
				return false
			}
		}
		return true

	case contiguousLocation:
		s, e := v.span()
		return rangeWithin(s, e, lower, upper)

	default:
		return false
	}
}

// LocationOverlap tests if the given location overlaps with the given bounds.
func LocationOverlap(loc Location, lower, upper int) bool {
	switch v := loc.(type) {
	case Complemented:
		return LocationOverlap(v.Location, lower, upper)

	case locationSlice:
		for _, l := range v.slice() {
			if LocationOverlap(l, lower, upper) {
				return true
			}
		}
		return false

	case contiguousLocation:
		s, e := v.span()
		return rangeOverlap(s, e, lower, upper)

	default:
		return false
	}
}

// Location represents a location in a sequence as defined by the INSDC feature
// table definition.
type Location interface {
	fmt.Stringer
	Len() int
	Region() Region
	Complement() Location
	Reverse(length int) Location
	Normalize(length int) Location
	Shift(i, n int) Location
	Expand(i, n int) Location
}

type contiguousLocation interface {
	Location
	span() (int, int)
}

type locationSlice interface {
	Location
	slice() []Location
}

// Locations represents a sortable list of locations.
type Locations []Location

// Len is the number of elements in the collection.
func (ll Locations) Len() int {
	return len(ll)
}

// Less reports whether the element with index i should sort before the element
// with index j.
func (ll Locations) Less(i, j int) bool {
	return LocationLess(ll[i], ll[j])
}

// Swap the elements with indexes i and j.
func (ll Locations) Swap(i, j int) {
	ll[i], ll[j] = ll[j], ll[i]
}

// Between represents a position between two bases. This will only make logical
// sense if the start and end positions are directly adjacent.
type Between int

func (between Between) span() (int, int) {
	pos := int(between)
	return pos, pos
}

// String satisfies the fmt.Stringer interface.
func (between Between) String() string {
	return fmt.Sprintf("%d^%d", between, between+1)
}

// Len returns the total length spanned by the location.
func (between Between) Len() int {
	return 0
}

// Region returns the region pointed to by the location.
func (between Between) Region() Region {
	head := int(between)
	return Segment{head, head}
}

// Complement returns the complement location.
func (between Between) Complement() Location {
	return Complemented{between}
}

// Reverse returns the reversed location for the given length sequence.
func (between Between) Reverse(length int) Location {
	return Between(length - 1 - int(between))
}

// Normalize returns a location normalized for the given length sequence.
func (between Between) Normalize(length int) Location {
	return Between(int(between) % length)
}

// Shift the location beyond the given position i by n.
func (between Between) Shift(i, n int) Location {
	return between.Expand(i, n)
}

// Expand the location beyond the given position i by n.
func (between Between) Expand(i, n int) Location {
	p := int(between)
	if i < p {
		p = Max(i, p+n)
	}
	return Between(p)
}

// Point represents a single base position in a sequence.
type Point int

func (point Point) span() (int, int) {
	p := int(point)
	return p, p + 1
}

// String satisfies the fmt.Stringer interface.
func (point Point) String() string {
	return strconv.Itoa(int(point + 1))
}

// Len returns the total length spanned by the location.
func (point Point) Len() int {
	return 1
}

// Region returns the region pointed to by the location.
func (point Point) Region() Region {
	head := int(point)
	return Segment{head, head + 1}
}

// Complement returns the complement location.
func (point Point) Complement() Location {
	return Complemented{point}
}

// Reverse returns the reversed location for the given length sequence.
func (point Point) Reverse(length int) Location {
	return Point(length - 1 - int(point))
}

// Normalize returns a location normalized for the given length sequence.
func (point Point) Normalize(length int) Location {
	return Point(int(point) % length)
}

// Shift the location beyond the given position i by n.
func (point Point) Shift(i, n int) Location {
	return point.Expand(i, n)
}

// Expand the location beyond the given position i by n.
func (point Point) Expand(i, n int) Location {
	p := int(point)
	if n < 0 && i == p {
		return Between(i)
	}
	if (0 <= n && i <= p) || (n < 0 && i < p) {
		p = Max(i, p+n)

	}
	return Point(p)
}

// Partial represents the partiality of a location range.
type Partial [2]bool

// Partiality values.
var (
	Complete    Partial = [2]bool{false, false}
	Partial5    Partial = [2]bool{true, false}
	Partial3    Partial = [2]bool{false, true}
	PartialBoth Partial = [2]bool{true, true}
)

// Ranged represents a contiguous region of bases in a sequence. The starting
// and ending positions of a Ranged may be partial.
type Ranged struct {
	Start   int
	End     int
	Partial Partial
}

func asComplete(loc Location) Location {
	switch v := loc.(type) {
	case Ranged:
		v.Partial = Complete
		return v
	case Joined:
		for i, u := range v {
			v[i] = asComplete(u)
		}
		return v
	case Ordered:
		for i, u := range v {
			v[i] = asComplete(u)
		}
		return v
	default:
		return v
	}
}

// PartialRange returns the range between the start and end positions where the
// specified ends are partial. They can be Complete, Partial5, Partial3, or
// PartialBoth.
func PartialRange(start, end int, partial Partial) Ranged {
	if end <= start {
		panic(fmt.Errorf("Ranged bounds out of range [%d:%d]", start, end))
	}
	/* DISCUSS: should a complete, one base range be reduced to a Point?
	if partial == Complete && start+1 == end {
		return Point(start)
	}
	*/
	return Ranged{start, end, partial}
}

// Range returns the complete range between the start and end positions.
func Range(start, end int) Ranged {
	return PartialRange(start, end, Complete)
}

func (ranged Ranged) span() (int, int) {
	return ranged.Start, ranged.End
}

// String satisfies the fmt.Stringer interface.
func (ranged Ranged) String() string {
	b := strings.Builder{}
	if ranged.Partial[0] {
		b.WriteByte('<')
	}
	b.WriteString(strconv.Itoa(ranged.Start + 1))
	b.WriteString("..")
	if ranged.Partial[1] {
		b.WriteByte('>')
	}
	b.WriteString(strconv.Itoa(ranged.End))
	return b.String()
}

// Len returns the total length spanned by the location.
func (ranged Ranged) Len() int {
	return ranged.End - ranged.Start
}

// Region returns the region pointed to by the location.
func (ranged Ranged) Region() Region {
	head, tail := ranged.Start, ranged.End
	return Segment{head, tail}
}

// Complement returns the complement location.
func (ranged Ranged) Complement() Location {
	return Complemented{ranged}
}

// Reverse returns the reversed location for the given length sequence.
func (ranged Ranged) Reverse(length int) Location {
	ret := PartialRange(length-ranged.End, length-ranged.Start, ranged.Partial)
	switch ret.Partial {
	case Partial5:
		ret.Partial = Partial3
	case Partial3:
		ret.Partial = Partial5
	}
	return ret
}

// Normalize returns a location normalized for the given length sequence.
func (ranged Ranged) Normalize(length int) Location {
	if ranged.Len() == length {
		return ranged.Expand(0, -ranged.Start)
	}
	start, end := ranged.Start%length, ranged.End%length
	if start < end {
		return PartialRange(start, end, ranged.Partial)
	}
	left, right := Range(start, length), Range(0, end)
	if ranged.Partial[0] {
		left.Partial = Partial5
	}
	if ranged.Partial[1] {
		right.Partial = Partial3
	}
	return Join(left, right)
}

// Shift the location beyond the given position i by n.
func (ranged Ranged) Shift(i, n int) Location {
	if n == 0 {
		return ranged
	}
	if n < 0 {
		return ranged.Expand(i, n)
	}
	start, end, partial := ranged.Start, ranged.End, ranged.Partial
	if start < i && i < end {
		left, right := Range(start, i), Range(i+n, end+n)
		if partial[0] {
			left.Partial = Partial5
		}
		if partial[1] {
			right.Partial = Partial3
		}
		return Join(left, right)
	}
	if i <= start {
		start += n
	}
	if i < end {
		end += n
	}
	return Ranged{start, end, partial}
}

// Expand the location beyond the given position i by n.
func (ranged Ranged) Expand(i, n int) Location {
	if n == 0 {
		return ranged
	}
	start, end, partial := ranged.Start, ranged.End, ranged.Partial
	if n < 0 {
		j := i - n
		if i <= start && start < j {
			partial[0] = true
		}
		if i < end && end <= j {
			partial[1] = true
		}
	}
	if (0 <= n && i <= start) || (n < 0 && i < start) {
		start = Max(i, start+n)
	}
	if (0 <= n && i < end) || (n < 0 && i <= end) {
		end = Max(i, end+n)
	}
	if start == end {
		return Between(start)
	}
	return Ranged{start, end, partial}
}

// Ambiguous represents a single base within a given range.
type Ambiguous [2]int

// String satisfies the fmt.Stringer interface.
func (ambiguous Ambiguous) String() string {
	return fmt.Sprintf("%d.%d", ambiguous[0]+1, ambiguous[1])
}

// Len returns the total length spanned by the location.
func (ambiguous Ambiguous) Len() int {
	return 1
}

func (ambiguous Ambiguous) span() (int, int) {
	return Unpack(ambiguous)
}

// Region returns the region pointed to by the location.
func (ambiguous Ambiguous) Region() Region {
	head, tail := Unpack(ambiguous)
	return Segment{head, tail}
}

// Complement returns the complement location.
func (ambiguous Ambiguous) Complement() Location {
	return Complemented{ambiguous}
}

// Reverse returns the reversed location for the given length sequence.
func (ambiguous Ambiguous) Reverse(length int) Location {
	return Ambiguous{length - ambiguous[1], length - ambiguous[0]}
}

// Normalize returns a location normalized for the given length sequence.
func (ambiguous Ambiguous) Normalize(length int) Location {
	return Ambiguous{ambiguous[0] % length, ambiguous[1] % length}
}

// Shift the location beyond the given position i by n.
func (ambiguous Ambiguous) Shift(i, n int) Location {
	if n == 0 {
		return ambiguous
	}
	if n < 0 {
		return ambiguous.Expand(i, n)
	}
	start, end := Unpack(ambiguous)
	if start < i && i < end {
		left, right := Ambiguous{start, i}, Ambiguous{i + n, end + n}
		return Order(left, right)
	}
	if i <= start {
		start += n
	}
	if i < end {
		end += n
	}
	return Ambiguous{start, end}
}

// Expand the location beyond the given position i by n.
func (ambiguous Ambiguous) Expand(i, n int) Location {
	if n == 0 {
		return ambiguous
	}
	start, end := Unpack(ambiguous)
	if (0 <= n && i <= start) || (n < 0 && i < start) {
		start = Max(i, start+n)
	}
	if (0 <= n && i < end) || (n < 0 && i <= end) {
		end = Max(i, end+n)
	}
	if start == end {
		return Between(start)
	}
	return Ambiguous{start, end}
}

// LocationList represents a singly linked list of Location objects.
type LocationList struct {
	Data Location
	Next *LocationList
}

// Len returns the length of the list.
func (ll *LocationList) Len() int {
	if ll.Next == nil {
		if ll.Data == nil {
			return 0
		}
		return 1
	}
	return ll.Next.Len() + 1
}

// Slice returns the slice representation of the list.
func (ll *LocationList) Slice() []Location {
	list := []Location{ll.Data}
	if ll.Next == nil {
		return list
	}
	return append(list, ll.Next.Slice()...)
}

// Push a Location object to the end of the list. If the Location object is
// equivalent to the last element, nothing happens. If the Location object can
// be joined with the last element to form a contiguous Location location, the
// last element will be replaced with the joined Location object. If the force
// option is false, then only partial ranges will be joined.
func (ll *LocationList) Push(loc Location, force bool) {
	if ll.Next != nil {
		ll.Next.Push(loc, force)
		return
	}

	if joined, ok := loc.(Joined); ok {
		for i := range joined {
			ll.Push(joined[i], force)
		}
		return
	}

	if ll.Data == nil {
		ll.Data = loc
		return
	}

	switch v := ll.Data.(type) {
	case Between:
		switch u := loc.(type) {
		case Between:
			if v == u {
				return
			}
		case Point:
			if int(v) == int(u) {
				ll.Data = u
				return
			}
		case Ranged:
			if int(v) == u.Start {
				ll.Data = u
				return
			}
		}

	case Point:
		switch u := loc.(type) {
		case Between:
			if int(v+1) == int(u) {
				return
			}
		case Point:
			if v == u {
				return
			}
		case Ranged:
			if int(v) == u.Start {
				ll.Data = u
				return
			}
		}

	case Ranged:
		switch u := loc.(type) {
		case Between:
			if v.End == int(u) {
				return
			}
		case Point:
			if v.End == int(u) {
				return
			}
		case Ranged:
			if ((v.Partial[1] && u.Partial[0]) || force) && v.End == u.Start {
				partial := Partial{v.Partial[0], u.Partial[1]}
				ll.Data = Ranged{v.Start, u.End, partial}
				return
			}
		}

	case Complemented:
		if u, ok := loc.(Complemented); ok {
			tmp := LocationList{u.Location, nil}
			tmp.Push(v.Location, force)
			ll.Data = Complemented{Join(tmp.Slice()...)}
			return
		}
	}

	ll.Next = &LocationList{loc, nil}
}

// Joined represents a list of Location locations. It is strongly recommended
// this be constructed using the Join helper function to reduce the list of
// Location locations to the simplest representation.
type Joined []Location

// Join the given Location locations. Will panic if no argument is given. The
// locations will first be reduced to the simplest representation by merging
// adjacent identical locations and contiguous locations. If the resulting list
// of locations have only one element, the elemnt will be returuned. Otherwise,
// a Joined object will be returned.
func Join(locs ...Location) Location {
	list := LocationList{}
	for _, loc := range locs {
		list.Push(loc, true)
	}

	switch list.Len() {
	case 0:
		panic("Join without arguments is not allowed")
	case 1:
		return list.Data
	default:
		return Joined(list.Slice())
	}
}

func (joined Joined) slice() []Location {
	return joined
}

// String satisfies the fmt.Stringer interface.
func (joined Joined) String() string {
	tmp := make([]string, len(joined))
	for i, loc := range joined {
		tmp[i] = loc.String()
	}
	return fmt.Sprintf("join(%s)", strings.Join(tmp, ","))
}

// Len returns the total length spanned by the location.
func (joined Joined) Len() int {
	n := 0
	for _, loc := range joined {
		n += loc.Len()
	}
	return n
}

// Region returns the region pointed to by the location.
func (joined Joined) Region() Region {
	rr := make(Regions, len(joined))
	for i, l := range joined {
		rr[i] = l.Region()
	}
	return rr
}

// Complement returns the complement location.
func (joined Joined) Complement() Location {
	return Complemented{joined}
}

// Reverse returns the reversed location for the given length sequence.
func (joined Joined) Reverse(length int) Location {
	ll := make([]Location, len(joined))
	for l, r := 0, len(ll)-1; l < r; l, r = l+1, r-1 {
		ll[l], ll[r] = joined[r].Reverse(length), joined[l].Reverse(length)
	}
	return Join(ll...)
}

// Normalize returns a location normalized for the given length sequence.
func (joined Joined) Normalize(length int) Location {
	ll := make([]Location, len(joined))
	for i, l := range joined {
		ll[i] = l.Normalize(length)
	}
	return Join(ll...)
}

// Shift the location beyond the given position i by n.
func (joined Joined) Shift(i, n int) Location {
	locs := make([]Location, len(joined))
	for j, loc := range joined {
		locs[j] = loc.Shift(i, n)
	}
	return Join(locs...)
}

// Expand the location beyond the given position i by n.
func (joined Joined) Expand(i, n int) Location {
	locs := make([]Location, len(joined))
	for j, loc := range joined {
		locs[j] = loc.Expand(i, n)
	}
	return Join(locs...)
}

func locationDelimiter(state *pars.State, result *pars.Result) bool {
	state.Push()
	c, err := pars.Next(state)
	if err != nil {
		state.Pop()
		return false
	}
	if c != ',' {
		state.Pop()
		return false
	}
	state.Advance()
	c, err = pars.Next(state)
	for ascii.IsSpace(c) && err == nil {
		state.Advance()
		c, err = pars.Next(state)
	}
	state.Drop()
	return true
}

// Ordered represents multiple locations.
type Ordered []Location

func flattenLocations(locs []Location) []Location {
	list := []Location{}
	for i := range locs {
		switch loc := locs[i].(type) {
		case Ordered:
			list = append(list, flattenLocations([]Location(loc))...)
		default:
			list = append(list, loc)
		}
	}
	return list
}

func (ordered Ordered) slice() []Location {
	return ordered
}

// Order takes the given Locations and returns an Ordered containing the
// simplest form.
func Order(locs ...Location) Location {
	list := flattenLocations(locs)
	switch len(list) {
	case 0:
		panic("Order without arguments is not allowed")
	case 1:
		return list[0]
	default:
		return Ordered(list)
	}
}

// String satisfies the fmt.Stringer interface.
func (ordered Ordered) String() string {
	tmp := make([]string, len(ordered))
	for i, loc := range ordered {
		tmp[i] = loc.String()
	}
	return fmt.Sprintf("order(%s)", strings.Join(tmp, ","))
}

// Len returns the total length spanned by the location.
func (ordered Ordered) Len() int {
	n := 0
	for _, loc := range ordered {
		n += loc.Len()
	}
	return n
}

// Region returns the region pointed to by the location.
func (ordered Ordered) Region() Region {
	rr := make(Regions, len(ordered))
	for i, l := range ordered {
		rr[i] = l.Region()
	}
	return rr
}

// Complement returns the complement location.
func (ordered Ordered) Complement() Location {
	return Complemented{ordered}
}

// Reverse returns the reversed location for the given length sequence.
func (ordered Ordered) Reverse(length int) Location {
	ll := make([]Location, len(ordered))
	for l, r := 0, len(ll)-1; l < r; l, r = l+1, r-1 {
		ll[l], ll[r] = ordered[r].Reverse(length), ordered[l].Reverse(length)
	}
	return Order(ll...)
}

// Normalize returns a location normalized for the given length sequence.
func (ordered Ordered) Normalize(length int) Location {
	ll := make([]Location, len(ordered))
	for i, l := range ordered {
		ll[i] = l.Normalize(length)
	}
	return Order(ll...)
}

// Shift the location beyond the given position i by n.
func (ordered Ordered) Shift(i, n int) Location {
	locs := make([]Location, len(ordered))
	for j, loc := range ordered {
		locs[j] = loc.Shift(i, n)
	}
	return Order(locs...)
}

// Expand the location beyond the given position i by n.
func (ordered Ordered) Expand(i, n int) Location {
	locs := make([]Location, len(ordered))
	for j, loc := range ordered {
		locs[j] = loc.Expand(i, n)
	}
	return Order(locs...)
}

// Complemented represents a location complemented for the given molecule type.
type Complemented struct {
	Location Location
}

// String satisfies the fmt.Stringer interface.
func (complement Complemented) String() string {
	return fmt.Sprintf("complement(%s)", complement.Location)
}

// Len returns the total length spanned by the location.
func (complement Complemented) Len() int {
	return complement.Location.Len()
}

// Region returns the region pointed to by the location.
func (complement Complemented) Region() Region {
	return complement.Location.Region().Complement()
}

// Complement returns the complement location.
func (complement Complemented) Complement() Location {
	return complement.Location
}

// Reverse returns the reversed location for the given length sequence.
func (complement Complemented) Reverse(length int) Location {
	return Complemented{complement.Location.Reverse(length)}
}

// Normalize returns a location normalized for the given length sequence.
func (complement Complemented) Normalize(length int) Location {
	return Complemented{complement.Location.Normalize(length)}
}

// Shift the location beyond the given position i by n.
func (complement Complemented) Shift(i, n int) Location {
	return Complemented{complement.Location.Shift(i, n)}
}

// Expand the location beyond the given position i by n.
func (complement Complemented) Expand(i, n int) Location {
	return Complemented{complement.Location.Expand(i, n)}
}

func parseBetween(state *pars.State, result *pars.Result) error {
	state.Push()
	if err := pars.Int(state, result); err != nil {
		state.Pop()
		return err
	}
	start := result.Value.(int)
	c, err := pars.Next(state)
	if err != nil {
		state.Pop()
		return err
	}
	if c != '^' {
		err := pars.NewError("expected `^`", state.Position())
		state.Pop()
		return err
	}
	state.Advance()
	if err := pars.Int(state, result); err != nil {
		state.Pop()
		return err
	}
	end := result.Value.(int)
	if start+1 != end {
		return fmt.Errorf("%d^%d is not a valid location: coordinates should be adjacent", start, end)
	}
	result.SetValue(Between(start))
	state.Drop()
	return nil
}

var parsePoint = pars.Parser(pars.Int).Map(func(result *pars.Result) error {
	point := result.Value.(int)
	result.SetValue(Point(point - 1))
	return nil
})

func parseRange(state *pars.State, result *pars.Result) error {
	state.Push()
	c, err := pars.Next(state)
	if err != nil {
		state.Pop()
		return err
	}
	partial5 := false
	if c == '<' {
		partial5 = true
		state.Advance()
	}
	if err := pars.Int(state, result); err != nil {
		state.Pop()
		return err
	}
	start := result.Value.(int) - 1
	if err := state.Request(2); err != nil {
		state.Pop()
		return err
	}
	if !bytes.Equal(state.Buffer(), []byte("..")) {
		err := pars.NewError("expected `..`", state.Position())
		state.Pop()
		return err
	}
	state.Advance()
	c, err = pars.Next(state)
	partial3 := false
	if c == '>' {
		partial3 = true
		state.Advance()
	}
	if err := pars.Int(state, result); err != nil {
		state.Pop()
		return err
	}
	end := result.Value.(int)

	// Some legacy entries have the partial marker in the end.
	c, err = pars.Next(state)
	if err == nil && c == '>' {
		partial3 = true
		state.Advance()
	}
	result.SetValue(Ranged{start, end, [2]bool{partial5, partial3}})
	state.Drop()
	return nil
}

func multipleLocationParser(state *pars.State, result *pars.Result) error {
	state.Push()
	if err := parseLocation(state, result); err != nil {
		state.Pop()
		return err
	}
	locs := []Location{result.Value.(Location)}
	for locationDelimiter(state, result) {
		if err := parseLocation(state, result); err != nil {
			state.Pop()
			return err
		}
		locs = append(locs, result.Value.(Location))
	}
	result.SetValue(locs)
	state.Drop()
	return nil
}

func parseAmbiguous(state *pars.State, result *pars.Result) error {
	state.Push()
	if err := pars.Int(state, result); err != nil {
		state.Pop()
		return err
	}
	start := result.Value.(int) - 1
	c, err := pars.Next(state)
	if err != nil {
		state.Pop()
		return err
	}
	if c != '.' {
		err := pars.NewError("expected `.`", state.Position())
		state.Pop()
		return err
	}
	state.Advance()
	if err := pars.Int(state, result); err != nil {
		state.Pop()
		return err
	}
	end := result.Value.(int)
	result.SetValue(Ambiguous{start, end})
	state.Drop()
	return nil
}

func parseJoin(state *pars.State, result *pars.Result) error {
	state.Push()
	if err := state.Request(5); err != nil {
		state.Pop()
		return err
	}
	if !bytes.Equal(state.Buffer(), []byte("join(")) {
		err := pars.NewError("expected `join(`", state.Position())
		state.Pop()
		return err
	}
	state.Advance()
	if err := multipleLocationParser(state, result); err != nil {
		return err
	}
	c, err := pars.Next(state)
	if err != nil {
		state.Pop()
		return err
	}
	if c != ')' {
		err := pars.NewError("expected `)`", state.Position())
		state.Pop()
		return err
	}
	state.Advance()
	result.SetValue(Join(result.Value.([]Location)...))
	state.Drop()
	return nil
}

func parseOrder(state *pars.State, result *pars.Result) error {
	state.Push()
	if err := state.Request(6); err != nil {
		state.Pop()
		return err
	}
	if !bytes.Equal(state.Buffer(), []byte("order(")) {
		err := pars.NewError("expected `order(`", state.Position())
		state.Pop()
		return err
	}
	state.Advance()
	if err := multipleLocationParser(state, result); err != nil {
		return err
	}
	c, err := pars.Next(state)
	if err != nil {
		state.Pop()
		return err
	}
	if c != ')' {
		err := pars.NewError("expected `)`", state.Position())
		state.Pop()
		return err
	}
	state.Advance()
	result.SetValue(Order(result.Value.([]Location)...))
	state.Drop()
	return nil
}

func parseComplement(q interface{}) pars.Parser {
	parser := pars.AsParser(q)
	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		if err := state.Request(11); err != nil {
			state.Pop()
			return err
		}
		if !bytes.Equal(state.Buffer(), []byte("complement(")) {
			err := pars.NewError("expected `complement(`", state.Position())
			state.Pop()
			return err
		}
		state.Advance()
		if err := parser(state, result); err != nil {
			state.Pop()
			return err
		}
		c, err := pars.Next(state)
		if err != nil {
			state.Pop()
			return err
		}
		if c != ')' {
			err := pars.NewError("expected `)`", state.Position())
			state.Pop()
			return err
		}
		state.Advance()
		result.SetValue(result.Value.(Location).Complement())
		state.Drop()
		return nil
	}
}

var parseComplementDefault = parseComplement(&parseLocation)

var parseLocation pars.Parser

// AsLocation interprets the given string as a Location.
func AsLocation(s string) (Location, error) {
	result, err := parseLocation.Parse(pars.FromString(s))
	if err != nil {
		return nil, err
	}
	return result.Value.(Location), nil
}

func init() {
	parseLocation = pars.Any(
		parseRange,
		parseBetween,
		parseAmbiguous,
		parseComplementDefault,
		parseJoin,
		parseOrder,
		parsePoint,
	)
}
