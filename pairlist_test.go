package gts

import "testing"

func TestPairList(t *testing.T) {
	d := Dictionary{}
	d.Set("foo", "foo")
	equals(t, d.Get("foo"), []string{"foo"})
	d.Set("foo", "bar")
	equals(t, d.Get("foo"), []string{"bar"})
	d.Del("foo")
	equals(t, d.Get("foo"), []string{})
}
