package gts

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	deep "gopkg.in/go-test/deep.v1"
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
