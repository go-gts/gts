package diff

import (
	"fmt"
	"strings"
)

func max(i, j int) int {
	if j > i {
		return j
	}
	return i
}

func compare(i, j int) int {
	switch {
	case i < j:
		return -1
	case j < i:
		return 1
	default:
		return 0
	}
}

const (
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
	surrogateGap = surrogateMax - surrogateMin
)

// Equiv represents a line equivalence value.
type Equiv = rune

func toIndex(equiv Equiv) int {
	if equiv < surrogateMax {
		return int(equiv)
	}
	return int(equiv - surrogateGap)
}

func toEquiv(index int) Equiv {
	if index < surrogateMin {
		return rune(index)
	}
	return rune(index + surrogateGap)
}

// Op represents an edit operation.
type Op interface {
	String() string
}

// Common represents a shared equiv.
type Common Equiv

// String returns the pretiffied edit string representation.
func (op Common) String() string {
	return fmt.Sprintf("%c", rune(op))
}

// Insert represents an insert operation.
type Insert Equiv

// String returns the pretiffied edit string representation.
func (op Insert) String() string {
	return fmt.Sprintf("\x1b[32m%c\x1b[0m", rune(op))
}

// Delete represents an delete operation.
type Delete Equiv

// String returns the pretiffied edit string representation.
func (op Delete) String() string {
	return fmt.Sprintf("\x1b[31m%c\x1b[0m", rune(op))
}

// CommonLine represents a shared line.
type CommonLine string

// String returns the pretiffied edit string representation.
func (op CommonLine) String() string {
	return "|\t" + string(op)
}

// InsertLine represents a line insert operation.
type InsertLine string

// String returns the pretiffied edit string representation.
func (op InsertLine) String() string {
	return fmt.Sprintf("\x1b[32m+\t%s\x1b[0m", string(op))
}

// DeleteLine represents a line delete operation.
type DeleteLine string

// String returns the pretiffied edit string representation.
func (op DeleteLine) String() string {
	return fmt.Sprintf("\x1b[31m-\t%s\x1b[0m", string(op))
}

// Point represents a point in the edit graph.
type Point struct {
	X, Y int
}

// Route represents a point in the edit graph with an associated route.
type Route struct {
	X, Y, R int
}

// Context represents a diff context.
type Context struct {
	a, b    []rune
	m, n    int
	paths   []int
	routes  []Route
	reverse bool
}

func generate(n, v int) []int {
	p := make([]int, n)
	for i := range p {
		p[i] = v
	}
	return p
}

// NewContext returns a new Context.
func NewContext(a, b []rune) *Context {
	m, n := len(a), len(b)
	paths := generate(m+n+3, -1)
	routes := make([]Route, 0)
	reverse := m >= n
	if reverse {
		a, b = b, a
		m, n = n, m
	}
	return &Context{a, b, m, n, paths, routes, reverse}
}

func (ctx *Context) compute() []Op {
	fp := generate(ctx.m+ctx.n+3, -1)

	offset := ctx.m + 1
	delta := ctx.n - ctx.m
	for p := 0; ; p++ {
		for k := -p; k <= delta-1; k++ {
			fp[k+offset] = ctx.snake(k, fp[k-1+offset]+1, fp[k+1+offset], offset)
		}
		for k := delta + p; k >= delta+1; k-- {
			fp[k+offset] = ctx.snake(k, fp[k-1+offset]+1, fp[k+1+offset], offset)
		}
		fp[delta+offset] = ctx.snake(delta, fp[delta-1+offset]+1, fp[delta+1+offset], offset)
		if fp[delta+offset] >= ctx.n {
			break
		}
	}

	r := ctx.paths[delta+offset]
	epc := make([]Point, 0)
	for r != -1 {
		epc = append(epc, Point{ctx.routes[r].X, ctx.routes[r].Y})
		r = ctx.routes[r].R
	}

	x, y := 1, 1
	px, py := 0, 0
	ses := []Op{}
	for i := len(epc) - 1; i >= 0; i-- {
		for (px < epc[i].X) || (py < epc[i].Y) {
			switch compare(epc[i].Y-epc[i].X, py-px) {
			case 1:
				r := ctx.b[py]
				ses = append(ses, Insert(r))
				y++
				py++
			case -1:
				r := ctx.a[px]
				ses = append(ses, Delete(r))
				x++
				px++
			default:
				ses = append(ses, Common(ctx.a[px]))
				x++
				y++
				px++
				py++
			}
		}
	}

	if ctx.reverse {
		for i, op := range ses {
			switch equiv := op.(type) {
			case Insert:
				ses[i] = Delete(equiv)
			case Delete:
				ses[i] = Insert(equiv)
			}
		}
	}
	return ses
}

func (ctx *Context) snake(k, p, pp, offset int) int {
	r := 0

	if p > pp {
		r = ctx.paths[k-1+offset]
	} else {
		r = ctx.paths[k+1+offset]
	}

	y := max(p, pp)
	x := y - k

	for x < ctx.m && y < ctx.n && ctx.a[x] == ctx.b[y] {
		x++
		y++
	}

	ctx.paths[k+offset] = len(ctx.routes)
	ctx.routes = append(ctx.routes, Route{x, y, r})

	return y
}

// Diff returns the shortest edit sequence of two strings.
func Diff(s, t string) []Op {
	return NewContext([]rune(s), []rune(t)).compute()
}

// LineDiff returns the line diff of the two strings.
func LineDiff(s, t string) []Op {
	equivs := make(map[string]Equiv)
	lines := make([]string, 0)
	slines, tlines := strings.Split(s, "\n"), strings.Split(t, "\n")
	a, b := make([]rune, len(slines)), make([]rune, len(tlines))

	for i, line := range slines {
		equiv, ok := equivs[line]
		if !ok {
			equiv = toEquiv(len(lines))
			equivs[line] = equiv
			lines = append(lines, line)
		}
		a[i] = equiv

	}

	for i, line := range tlines {
		equiv, ok := equivs[line]
		if !ok {
			equiv = toEquiv(len(lines))
			equivs[line] = equiv
			lines = append(lines, line)
		}
		b[i] = equiv
	}

	ops := Diff(string(a), string(b))
	ses := make([]Op, len(ops))
	for i, op := range ops {
		switch equiv := op.(type) {
		case Common:
			index := toIndex(Equiv(equiv))
			ses[i] = CommonLine(lines[index])
		case Insert:
			index := toIndex(Equiv(equiv))
			ses[i] = InsertLine(lines[index])
		case Delete:
			index := toIndex(Equiv(equiv))
			ses[i] = DeleteLine(lines[index])
		}
	}
	return ses
}
