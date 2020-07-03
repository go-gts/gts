package seqio

import (
	"testing"

	"github.com/go-gts/gts/testutils"
)

func TestPairList(t *testing.T) {
	d := Dictionary{}
	d.Set("foo", "foo")
	testutils.Equals(t, d.Get("foo"), []string{"foo"})
	d.Set("foo", "bar")
	testutils.Equals(t, d.Get("foo"), []string{"bar"})
	d.Del("foo")
	testutils.Equals(t, d.Get("foo"), []string{})
}
