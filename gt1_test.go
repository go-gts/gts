package gt1_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ktnyt/pars"
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
