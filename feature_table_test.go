package gts

import (
	"sort"
	"strings"
	"testing"

	pars "gopkg.in/ktnyt/pars.v2"
)

func TestFeatureTableIO(t *testing.T) {
	in := ReadGolden(t)

	parser := pars.Exact(FeatureTableParser(""))
	state := pars.FromString(in)
	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("while parsing`\n%s\n`: %v", in, err)
	}

	switch ft := result.Value.(type) {
	case FeatureTable:
		b := strings.Builder{}
		n, err := ft.Format("     ", 21).WriteTo(&b)
		if err != nil {
			t.Errorf("qf.WriteTo(w) = %d, %v, want %d, nil", n, err, n)
		}
		out := b.String()
		equals(t, out, in)

		cp := FeatureTable{}
		for _, f := range ft.Features {
			cp.Add(f)
		}
		differs(t, cp, ft)
		sort.Sort(ft)
		equals(t, cp, ft)

		f := NewFeature("source", NewRangeLocation(39, 42), Values{})
		cp.Add(f)
		ft.Insert(len(ft.Features)/2, f)
		differs(t, cp, ft)
		sort.Sort(ft)
		equals(t, cp, ft)
	default:
		t.Errorf("result.Value.(type) = %T, want %T", ft, FeatureTable{})
	}
}
