package gts

import (
	"strings"
	"testing"

	pars "gopkg.in/ktnyt/pars.v2.4"
)

func TestFeatureIO(t *testing.T) {
	s := ReadGolden(t)
	ss := RecordSplit(s)

	for _, in := range ss {
		state := pars.FromString(in)
		parser := pars.Exact(FeatureParser(""))
		result, err := parser.Parse(state)
		if err != nil {
			t.Errorf("while parsing`\n%s\n`: %v", in, err)
			return
		}
		switch f := result.Value.(type) {
		case Feature:
			b := strings.Builder{}
			n, err := f.Format("     ", 21).WriteTo(&b)
			if err != nil {
				t.Errorf("qf.WriteTo(w) = %d, %v, want %d, nil", n, err, n)
			}
			out := b.String()
			if out != in {
				t.Errorf("f.Format(%q, 21) = %q, want %q", "     ", out, in)
			}
			switch f.Key {
			case "CDS":
				s := f.Qualifiers.Get("translation")[0]
				seq := Seq(strings.ReplaceAll(s, "\n", ""))
				out := f.Translation()
				equals(t, out, seq)
			default:
				out := f.Translation()
				equals(t, out, Sequence(nil))
			}
		default:
			t.Errorf("result.Value.(type) = %T, want %T", f, Feature{})
		}
	}

	malformedKeyline := "     source          \n"

	for _, in := range []string{malformedKeyline} {
		state := pars.FromString(in)
		parser := pars.Exact(FeatureParser(""))
		if _, err := parser.Parse(state); err == nil {
			t.Errorf("while parsing`\n%s\n`: expected error", in)
		}
	}
}
