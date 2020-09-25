package utils

import "testing"

func TestUnpack(t *testing.T) {
	p := [2]int{0, 1}
	x, y := Unpack(p)
	if x != p[0] || y != p[1] {
		t.Errorf("Unpack(%v) = (%d, %d), want (%d, %d)", p, x, y, p[0], p[1])
	}
}

func TestMin(t *testing.T) {
	a, b := 1, 2
	if Min(a, b) != a {
		t.Errorf("Min(%d, %d) = %d, want %d", a, b, b, a)
	}
	if Min(b, a) != a {
		t.Errorf("Min(%d, %d) = %d, want %d", a, b, b, a)
	}
}

func TestCmp(t *testing.T) {
	a, b := 1, 2
	if Compare(a, b) != -1 {
		t.Errorf("cmp(%d, %d) = %d, want %d", a, b, Compare(a, b), -1)
	}
	if Compare(b, a) != 1 {
		t.Errorf("cmp(%d, %d) = %d, want %d", b, a, Compare(b, a), 1)
	}
	if Compare(a, a) != 0 {
		t.Errorf("cmp(%d, %d) = %d, want %d", a, a, Compare(a, a), 0)
	}
}
