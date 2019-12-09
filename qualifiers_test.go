package gts

import (
	"testing"
)

func TestQualifiers(t *testing.T) {
	qs := Qualifiers{}

	equals(t, Qualifiers(nil).Get("foo") == nil, true)
	equals(t, qs.Get("foo") == nil, true)
	qs.Set("foo", "bar")
	equals(t, qs.Get("foo"), []string{"bar"})
	qs.Add("foo", "baz")
	equals(t, qs.Get("foo"), []string{"bar", "baz"})
	qs.Del("foo")
	equals(t, qs.Get("foo") == nil, true)
}
