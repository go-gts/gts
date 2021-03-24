package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func TestProps(t *testing.T) {
	p := Props{}
	testutils.Equals(t, p.Get("foo") == nil, true)
	p.Add("foo", "baz")
	testutils.Equals(t, p.Get("foo"), []string{"baz"})
	testutils.Equals(t, p.Has("foo"), true)
	p.Set("foo", "bar")
	testutils.Equals(t, p.Get("foo"), []string{"bar"})
	p.Add("foo", "baz")
	testutils.Equals(t, p.Get("foo"), []string{"bar", "baz"})
	testutils.Equals(t, p.Keys(), []string{"foo"})
	testutils.Equals(t, p.Items(), []Item{{"foo", "bar"}, {"foo", "baz"}})
	p.Del("foo")
	testutils.Equals(t, p.Get("foo") == nil, true)
	testutils.Equals(t, p.Has("foo"), false)
}
