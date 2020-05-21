package utils

import "testing"

func TestMin(t *testing.T) {
	a, b := 1, 2
	if Min(a, b) != a {
		t.Errorf("min(%d, %d) = %d, want %d", a, b, b, a)
	}
	if Min(b, a) != a {
		t.Errorf("min(%d, %d) = %d, want %d", a, b, b, a)
	}
}
