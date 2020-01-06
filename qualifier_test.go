package gts

import (
	"strings"
	"testing"

	pars "gopkg.in/pars.v2"
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
			b := strings.Builder{}
			n, err := q.Format(prefix).WriteTo(&b)
			if err != nil {
				t.Errorf("qf.WriteTo(w) = %d, %v, want %d, nil", n, err, n)
			}
			out := b.String()
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

	PanicTest(t, func() { Qualifier{"foo", "bar"}.Format("").String() })
}

func TestQualifierListIO(t *testing.T) {
	prefix := strings.Repeat(" ", 21)

	in := ReadGolden(t)
	state := pars.FromString(in)
	singleParser := pars.Seq(QualifierParser(prefix), pars.EOL)
	parser := pars.Exact(pars.Many(singleParser))
	_, err := parser.Parse(state)
	if err != nil {
		t.Errorf("while parsing`\n%s\n`: %v", in, err)
		return
	}
}
