package gts

import (
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

var topologyTests = []Topology{
	Linear,
	Circular,
}

func TestTopology(t *testing.T) {
	for _, in := range topologyTests {
		s := in.String()
		out, err := AsTopology(s)
		if err != nil {
			t.Errorf("AsTopology(%q): %v", s, err)
		}
		if in != out {
			t.Errorf("AsTopology(%q) = %q, expected %q", in.String(), out.String(), in.String())
		}
	}
}

func TestTopologyFail(t *testing.T) {
	in := ""
	out, err := AsTopology(in)
	if err == nil {
		t.Errorf("expected error in AsTopology(%q)", in)
	}
	if out.String() != "" {
		t.Errorf("Topology(%d).String() = %q, expected %q", out, out.String(), in)
	}
}

func (wt withTest) WithTopology(t Topology) Sequence {
	if _, ok := wt.info.(Topology); ok {
		return wt.WithInfo(t)
	}
	return wt
}

var withTopologyTests = []struct {
	in  Sequence
	out Sequence
}{
	{New(nil, nil, nil), New(nil, nil, nil)},
	{newWithTest(nil, nil, nil), newWithTest(nil, nil, nil)},
	{newWithTest(Linear, nil, nil), newWithTest(Circular, nil, nil)},
}

func TestWithTopology(t *testing.T) {
	for _, tt := range withTopologyTests {
		out := WithTopology(tt.in, Circular)
		testutils.Equals(t, out, tt.out)
	}
}
