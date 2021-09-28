package seqio

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

var multiLineStringTests = []struct {
	in  []string
	out string
}{
	{[]string{"foo"}, "foo"},             // case 1
	{[]string{"foo", "bar"}, "foo\nbar"}, // case 2
}

func TestMultiLineString(t *testing.T) {
	for i, tt := range multiLineStringTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := multiLineString(tt.in...)
			testutils.Diff(t, out, tt.out)
		})
	}
}

var addPrefixTests = []struct {
	in, out string
}{
	{"foo", "foo"},               // case 1
	{"foo\nbar", "foo\n    bar"}, // case 2
}

func TestAddPrefix(t *testing.T) {
	for i, tt := range addPrefixTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			out := AddPrefix(tt.in, "    ")
			testutils.DiffLine(t, out, tt.out)
		})
	}
}
