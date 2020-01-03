package gts

import (
	"testing"
)

func TestValues(t *testing.T) {
	v := Values{}
	equals(t, Values(nil).Get("foo") == nil, true)
	equals(t, v.Get("foo") == nil, true)
	v.Set("foo", "bar")
	equals(t, v.Get("foo"), []string{"bar"})
	v.Add("foo", "baz")
	equals(t, v.Get("foo"), []string{"bar", "baz"})
	v.Del("foo")
	equals(t, v.Get("foo") == nil, true)
}
