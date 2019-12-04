package gt1_test

import (
	"strings"
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

func testQualifierIOValid(s string) assert.F {
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

func testQualifierIOInvalid(s string) assert.F {
	prefix := strings.Repeat(" ", 21)
	state := pars.FromString(s)
	parser := pars.Exact(gt1.QualifierParser(prefix))
	return assert.IsError(parser(state, pars.Void))
}

func TestQualifierIO(t *testing.T) {
	s := ReadGolden(t)
	ss := RecordSplit(s)

	n := len(ss) - 2
	validStrings, invalidStrings := ss[:n], ss[n:]

	validCases := make([]assert.F, len(validStrings))
	for i, s := range validStrings {
		validCases[i] = testQualifierIOValid(s)
	}

	invalidCases := make([]assert.F, len(invalidStrings))
	for i, s := range invalidStrings {
		invalidCases[i] = testQualifierIOInvalid(s)
	}

	assert.Apply(t,
		assert.C("valid", validCases...),
		assert.C("invalid", invalidCases...),
	)
}
