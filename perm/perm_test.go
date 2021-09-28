package perm

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

var permTests = []struct {
	in    Interface
	perms []Interface
}{
	{
		IntSlice{0, 1, 2},
		[]Interface{
			IntSlice{0, 1, 2},
			IntSlice{1, 0, 2},
			IntSlice{2, 0, 1},
			IntSlice{0, 2, 1},
			IntSlice{1, 2, 0},
			IntSlice{2, 1, 0},
		},
	},
	{
		Float64Slice{math.E, math.Pi, math.Phi},
		[]Interface{
			Float64Slice{math.E, math.Pi, math.Phi},
			Float64Slice{math.Pi, math.E, math.Phi},
			Float64Slice{math.Phi, math.E, math.Pi},
			Float64Slice{math.E, math.Phi, math.Pi},
			Float64Slice{math.Pi, math.Phi, math.E},
			Float64Slice{math.Phi, math.Pi, math.E},
		},
	},
	{
		StringSlice{"foo", "bar", "baz"},
		[]Interface{
			StringSlice{"foo", "bar", "baz"},
			StringSlice{"bar", "foo", "baz"},
			StringSlice{"baz", "foo", "bar"},
			StringSlice{"foo", "baz", "bar"},
			StringSlice{"bar", "baz", "foo"},
			StringSlice{"baz", "bar", "foo"},
		},
	},
}

func TestPerm(t *testing.T) {
	for i, tt := range permTests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			perm := Permutate(tt.in)
			for j, exp := range tt.perms {
				if ok := perm.Next(); !ok {
					t.Errorf("returned false on permutation %d", j)
				}

				if !reflect.DeepEqual(tt.in, exp) {
					t.Errorf("permutation %d = %v, want %v", j, tt.in, exp)
				}
			}

			if perm.Next() {
				t.Errorf("permutation should be exhausted")
			}
		})
	}
}
