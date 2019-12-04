package gt1_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

func testFeatureIOStrings(s string) assert.F {
	state := pars.FromString(s)
	result := pars.Result{}
	parser := pars.Exact(gt1.FeatureParser(""))

	err := parser(state, &result)
	feature, ok := result.Value.(gt1.Feature)

	return assert.All(
		assert.NoError(err),
		assert.True(ok),
		assert.Equal(feature.Format("     ", 21), s),
	)
}

func TestFeatureIO(t *testing.T) {
	s := ReadGolden(t)
	testFeatureStrings := RecordSplit(s)
	cases := make([]assert.F, len(testFeatureStrings))

	for i, s := range testFeatureStrings {
		cases[i] = testFeatureIOStrings(s)
	}

	assert.Apply(t, cases...)
}
