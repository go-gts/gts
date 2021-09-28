package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func TestUnpack(t *testing.T) {
	a, b := 39, 42
	p := [2]int{a, b}
	x, y := Unpack(p)
	if a != x || b != y {
		t.Errorf("Unpack(%v) = (%d, %d), want (%d, %d)", p, x, y, a, b)
	}
}

var absTests = [][2]int{
	{-42, 42},
	{42, 42},
}

func TestAbs(t *testing.T) {
	for i, tt := range absTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			in, exp := Unpack(tt)
			out := Abs(in)
			if out != exp {
				t.Errorf("Abs(%d) = %d, want %d", in, out, exp)
			}
		})
	}
}

var compareTests = []struct {
	i, j int
	out  int
}{
	{39, 42, -1}, // case 1
	{42, 39, 1},  // case 2
	{42, 42, 0},  // case 3
}

func TestCompare(t *testing.T) {
	for i, tt := range compareTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := Compare(tt.i, tt.j)
			if out != tt.out {
				t.Errorf("Compare(%d, %d) = %d, want %d", tt.i, tt.j, out, tt.out)
			}
		})
	}
}

var minTests = []struct {
	i, j int
	out  int
}{
	{39, 42, 39}, // case 1
	{42, 39, 39}, // case 2
}

func TestMin(t *testing.T) {
	for i, tt := range minTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := Min(tt.i, tt.j)
			if out != tt.out {
				t.Errorf("Min(%d, %d) = %d, want %d", tt.i, tt.j, out, tt.out)
			}
		})
	}
}

var maxTests = []struct {
	i, j int
	out  int
}{
	{39, 42, 42}, // case 1
	{42, 39, 42}, // case 2
}

func TestMax(t *testing.T) {
	for i, tt := range maxTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := Max(tt.i, tt.j)
			if out != tt.out {
				t.Errorf("Max(%d, %d) = %d, want %d", tt.i, tt.j, out, tt.out)
			}
		})
	}
}
