package testutils

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ReadGolden will attempt to read the golden file associated to the test.
func ReadGolden(t *testing.T) string {
	t.Helper()
	p, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+".golden"))
	if err != nil {
		t.Fatalf("failed to read .golden file: %s", err)
	}
	return string(p)
}

// ReadTestfile will open a file in the testdata directory.
func ReadTestfile(t *testing.T, path string) string {
	t.Helper()
	p, err := ioutil.ReadFile(filepath.Join("testdata", path))
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	return string(p)
}

// Equals checks the equality of two objects using go-test/deep.
func Equals(t *testing.T, a, b interface{}) {
	t.Helper()
	if diff := deep.Equal(a, b); diff != nil {
		t.Error(diff)
	}
}

// Diff checks the equality of two strings and reports its diff if they differ.
func Diff(t *testing.T, a, b string) {
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

// Panics will test if the given function panics.
func Panics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if recover() == nil {
			t.Errorf("given function did not panic")
		}
	}()
	f()
}
