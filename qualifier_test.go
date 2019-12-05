package gts_test

import (
	"strings"
	"testing"

	"gopkg.in/ktnyt/assert.v1"
	"gopkg.in/ktnyt/gts.v0"
	"gopkg.in/ktnyt/pars.v2"
)

func testQualifierIOValid(s string) assert.F {
	prefix := strings.Repeat(" ", 21)

	state := pars.FromString(s)
	result := pars.Result{}
	parser := pars.Exact(gts.QualifierParser(prefix))

	err := parser(state, &result)
	q, ok := result.Value.(gts.Qualifier)

	return assert.All(
		assert.NoError(err),
		assert.True(ok),
		assert.Equal(q.Format(prefix), s),
	)
}

func testQualifierIOInvalid(s string) assert.F {
	prefix := strings.Repeat(" ", 21)
	state := pars.FromString(s)
	parser := pars.Exact(gts.QualifierParser(prefix))
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
