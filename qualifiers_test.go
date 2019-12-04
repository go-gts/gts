package gt1_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
)

func TestQualifiers(t *testing.T) {
	qs := gt1.Qualifiers{}

	assert.Apply(t,
		assert.True(gt1.Qualifiers(nil).Get("foo") == nil),
		assert.True(qs.Get("foo") == nil),
		assert.Eval(func() { qs.Set("foo", "bar") }),
		assert.Equal(qs.Get("foo"), []string{"bar"}),
		assert.Eval(func() { qs.Add("foo", "baz") }),
		assert.Equal(qs.Get("foo"), []string{"bar", "baz"}),
		assert.Eval(func() { qs.Del("foo") }),
		assert.True(qs.Get("foo") == nil),
	)
}
