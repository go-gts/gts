package seqio

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

var flatfileSplitTests = []struct {
	in  string
	out []string
}{
	{".", nil},
	{"foo; bar.", []string{"foo", "bar"}},
}

func TestFlatfileSplit(t *testing.T) {
	for _, tt := range flatfileSplitTests {
		out := FlatFileSplit(tt.in)
		testutils.Equals(t, out, tt.out)
	}
}

var addPrefixTests = []struct {
	in, out string
}{
	{"foo", "foo"},
	{"foo\nbar", "foo\n    bar"},
}

func TestAddPrefix(t *testing.T) {
	for _, tt := range addPrefixTests {
		out := AddPrefix(tt.in, "    ")
		testutils.DiffLine(t, out, tt.out)
	}
}
