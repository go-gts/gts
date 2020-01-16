package gts

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	deep "gopkg.in/go-test/deep.v1"
	pars "gopkg.in/pars.v2"
)

func ReadGolden(t *testing.T) string {
	if h, ok := testing.TB(t).(interface{ Helper() }); ok {
		h.Helper()
	}
	p, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+".golden"))
	if err != nil {
		t.Fatalf("failed to read .golden file: %s", err)
	}
	return string(p)
}

func RecordSplit(s string) []string {
	term := pars.Seq("\n//", pars.EOL)
	parser := pars.Many(pars.Seq(pars.Until(term), term).Child(0))
	state := pars.FromString(s)
	result := pars.Result{}
	if err := parser(state, &result); err != nil {
		panic(err)
	}
	ss := make([]string, len(result.Children))
	for i, child := range result.Children {
		ss[i] = string(child.Token)
	}
	return ss
}

func PanicTest(t *testing.T, f func()) {
	defer func() {
		if recover() == nil {
			t.Errorf("function did not panic")
		}
	}()
	f()
}

func same(a, b interface{}) bool { return deep.Equal(a, b) == nil }

func equals(t *testing.T, a, b interface{}) {
	if h, ok := testing.TB(t).(interface{ Helper() }); ok {
		h.Helper()
	}
	if diff := deep.Equal(a, b); diff != nil {
		t.Error(diff)
	}
}

func differs(t *testing.T, a, b interface{}) {
	if h, ok := testing.TB(t).(interface{ Helper() }); ok {
		h.Helper()
	}
	if deep.Equal(a, b) == nil {
		t.Errorf("\nexpected: %v\n  actual: %v\nto be different", a, b)
	}
}
