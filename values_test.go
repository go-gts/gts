package gts

import (
	"testing"

	"github.com/go-gts/gts/testutils"
)

func TestValues(t *testing.T) {
	v := Values{}
	testutils.Equals(t, Values(nil).Get("foo") == nil, true)
	testutils.Equals(t, v.Get("foo") == nil, true)
	v.Set("foo", "bar")
	testutils.Equals(t, v.Get("foo"), []string{"bar"})
	v.Add("foo", "baz")
	testutils.Equals(t, v.Get("foo"), []string{"bar", "baz"})
	v.Del("foo")
	testutils.Equals(t, v.Get("foo") == nil, true)
}
