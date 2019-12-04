package gt1_test

import (
	"strings"
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

func testQualifierIOString(s string) assert.F {
	prefix := strings.Repeat(" ", 21)

	state := pars.FromString(s)
	result := pars.Result{}
	parser := pars.Exact(gt1.QualifierParser(prefix))

	err := parser(state, &result)
	q, ok := result.Value.(gt1.Qualifier)

	return assert.All(
		assert.NoError(err),
		assert.True(ok),
		assert.Equal(q.Format(prefix), s),
	)
}

func TestQualifierIO(t *testing.T) {
	s := ReadGolden(t)
	ss := RecordSplit(s)
	cases := make([]assert.F, len(ss))
	for i, s := range ss {
		cases[i] = testQualifierIOString(s)
	}
	assert.Apply(t, cases...)
}
