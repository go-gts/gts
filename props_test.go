package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func TestProps(t *testing.T) {
	p := Props{}
	testutils.Equals(t, p.Len(), 0)
	testutils.Equals(t, p.Get("foo") == nil, true)

	p.Add("foo", "baz")
	testutils.Equals(t, p.Len(), 1)
	testutils.Equals(t, p.Get("foo"), []string{"baz"})
	testutils.Equals(t, p.Has("foo"), true)

	p.Set("foo", "bar")
	testutils.Equals(t, p.Len(), 1)
	testutils.Equals(t, p.Get("foo"), []string{"bar"})

	q := p.Clone()

	p.Add("foo", "baz")
	testutils.Equals(t, p.Len(), 2)
	testutils.Equals(t, p.Get("foo"), []string{"bar", "baz"})
	testutils.Equals(t, p.Keys(), []string{"foo"})
	testutils.Equals(t, p.Items(), []Item{{"foo", "bar"}, {"foo", "baz"}})

	p.Del("foo")
	testutils.Equals(t, p.Len(), 0)
	testutils.Equals(t, p.Get("foo") == nil, true)
	testutils.Equals(t, p.Has("foo"), false)

	p.Set("foo", "bar")
	testutils.Equals(t, q.Len(), 1)
	testutils.Equals(t, q.Get("foo"), []string{"bar"})
}
