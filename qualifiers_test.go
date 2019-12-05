package gts_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gts"
)

func TestQualifiers(t *testing.T) {
	qs := gts.Qualifiers{}

	assert.Apply(t,
		assert.True(gts.Qualifiers(nil).Get("foo") == nil),
		assert.True(qs.Get("foo") == nil),
		assert.Eval(func() { qs.Set("foo", "bar") }),
		assert.Equal(qs.Get("foo"), []string{"bar"}),
		assert.Eval(func() { qs.Add("foo", "baz") }),
		assert.Equal(qs.Get("foo"), []string{"bar", "baz"}),
		assert.Eval(func() { qs.Del("foo") }),
		assert.True(qs.Get("foo") == nil),
	)
}
