package gts

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func ReadGolden(t *testing.T) string {
	t.Helper()
	p, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+".golden"))
	if err != nil {
		t.Fatalf("failed to read .golden file: %s", err)
	}
	return string(p)
}

func equals(t *testing.T, a, b interface{}) {
	t.Helper()
	if diff := deep.Equal(a, b); diff != nil {
		t.Error(diff)
	}
}

func diff(t *testing.T, a, b string) {
	t.Helper()
	if a != b {
		dmp := diffmatchpatch.New()
		lineText1, lineText2, array := dmp.DiffLinesToChars(a, b)
		diffs := dmp.DiffMain(lineText1, lineText2, false)
		if len(diffs) > 0 {
			lineDiffs := dmp.DiffCharsToLines(diffs, array)
			t.Errorf("\n%s", dmp.DiffPrettyText(lineDiffs))
		}
	}
}

func panics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if recover() == nil {
			t.Errorf("given function did not panic")
		}
	}()
	f()
}
