package gts

import (
	"strings"
	"testing"

	pars "gopkg.in/ktnyt/pars.v2"
)

func TestQualifierIO(t *testing.T) {
	prefix := strings.Repeat(" ", 21)

	s := ReadGolden(t)
	ss := RecordSplit(s)

	for _, in := range ss {
		state := pars.FromString(in)
		parser := pars.Exact(QualifierParser(prefix))
		result, err := parser.Parse(state)
		if err != nil {
			t.Errorf("while parsing`\n%s\n`: %v", in, err)
			return
		}
		switch q := result.Value.(type) {
		case Qualifier:
			out := q.Format(prefix)
			if out != in {
				t.Errorf("q.Format(%q) = %q, want %q", prefix, out, in)
			}
		default:
			t.Errorf("result.Value.(type) = %T, want %T", q, Qualifier{})
		}
	}

	for _, in := range []string{"/sex=female", "/pseudo=\"true\""} {
		state := pars.FromString(in)
		parser := pars.Exact(QualifierParser(""))
		_, err := parser.Parse(state)
		if err == nil {
			t.Errorf("while parsing`\n%s\n`: expected error", in)
		}
	}

	PanicTest(t, func(t *testing.T) {
		t.Helper()
		Qualifier{"foo", "bar"}.Format("")
	})
}
