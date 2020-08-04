package gts

import (
	"testing"
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
