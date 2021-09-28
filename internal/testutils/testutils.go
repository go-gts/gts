package testutils

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-gts/gts/internal/diff"
	"github.com/go-test/deep"
)

// RunCase wraps the test case with a subtest with the given index.
func RunCase(t *testing.T, i int, f func(t *testing.T)) {
	t.Helper()
	t.Run(fmt.Sprintf("case %d", i+1), f)
}

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

// ReadTestfilePkg will open a file in the testdata directory of the gievn pkg.
func ReadTestfilePkg(t *testing.T, path, pkg string) string {
	t.Helper()
	p, err := ioutil.ReadFile(filepath.Join(pkg, "testdata", path))
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

// Differs checks the equality of two objects.
func Differs(t *testing.T, a, b interface{}) {
	t.Helper()
	if diff := deep.Equal(a, b); diff == nil {
		t.Errorf("expected %v != %v", a, b)
	}
}

func escape(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

// Diff checks the equality of two strings and reports its diff if they differ.
func Diff(t *testing.T, a, b string) {
	t.Helper()
	if a != b {
		ops := diff.Diff(escape(a), escape(b))
		ss := make([]string, len(ops))
		for i, op := range ops {
			ss[i] = op.String()
		}
		s := strings.Join(ss, "")
		t.Errorf("\n%s", s)
	}
}

// DiffLine checks the equality of two strings and reports its diff by lines
// if they differ.
func DiffLine(t *testing.T, a, b string) {
	t.Helper()
	if a != b {
		ops := diff.LineDiff(a, b)
		lines := make([]string, len(ops))
		for i, op := range ops {
			lines[i] = op.String()
		}
		s := strings.Join(lines, "\n")
		t.Errorf("\n%s", s)
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
