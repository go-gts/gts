package perm

// An implementation of Interface can be permutated by the routines in this
// package. The methods refer to elements of the underlying collection by
// integer index.
type Interface interface {
	// Len is the number of elements in the collection.
	Len() int

	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
}

// IntSlice attaches the methods of Interface to []int.
type IntSlice []int

// Len is the number of elements in the collection.
func (x IntSlice) Len() int {
	return len(x)
}

// Swap swaps the elements with indexes i and j.
func (x IntSlice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

// Float64Slice attaches the methods of Interface to []float64.
type Float64Slice []float64

// Len is the number of elements in the collection.
func (x Float64Slice) Len() int {
	return len(x)
}

// Swap swaps the elements with indexes i and j.
func (x Float64Slice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

// StringSlice attaches the methods of Interface to []string.
type StringSlice []string

// Len is the number of elements in the collection.
func (x StringSlice) Len() int {
	return len(x)
}

// Swap swaps the elements with indexes i and j.
func (x StringSlice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

// Permutator provides iterative access to the permutations of a collection
// which satisfies Interface.
type Permutator struct {
	x Interface
	p []int
	i int
}

// Permutate returns a permutator for the given collection.
func Permutate(data Interface) *Permutator {
	p := make([]int, data.Len()+1)
	for i := range p {
		p[i] = i
	}
	return &Permutator{data, p, 0}
}

// Next reports if there are any possible unexplored permutations left and
// permutes the collection to the next permutation if there is.
func (perm *Permutator) Next() bool {
	if perm.i == 0 {
		perm.i = 1
		return true
	}

	x, p, i := perm.x, perm.p, perm.i
	ok := i < x.Len()

	if ok {
		p[i]--
		j := 0
		if i&1 == 1 {
			j = p[i]
		}
		x.Swap(i, j)
		for i = 1; p[i] == 0; i++ {
			p[i] = i
		}
	}

	perm.i = i
	return ok
}
