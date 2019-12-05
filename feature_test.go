package gts_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gts"
	"github.com/ktnyt/pars"
)

func testFeatureIOStrings(s string) assert.F {
	state := pars.FromString(s)
	result := pars.Result{}
	parser := pars.Exact(gts.FeatureParser(""))

	err := parser(state, &result)
	feature, ok := result.Value.(gts.Feature)

	return assert.All(
		assert.NoError(err),
		assert.True(ok),
		assert.Equal(feature.Format("     ", 21), s),
	)
}

func TestFeatureIO(t *testing.T) {
	s := ReadGolden(t)
	ss := RecordSplit(s)
	cases := make([]assert.F, len(ss))
	for i, s := range ss {
		cases[i] = testFeatureIOStrings(s)
	}
	assert.Apply(t, cases...)
}
