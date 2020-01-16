package gts

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	pars "gopkg.in/pars.v2"
	msgpack "gopkg.in/vmihailenco/msgpack.v4"
	yaml "gopkg.in/yaml.v3"
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
}

func TestQualifierEncoding(t *testing.T) {
	in := &Qualifier{"foo", "bar"}

	t.Run("JSON", func(t *testing.T) {
		out := &Qualifier{}
		rw := &bytes.Buffer{}
		enc := json.NewEncoder(rw)
		if err := enc.Encode(in); err != nil {
			t.Errorf("enc.Encode(in): %v", err)
			return
		}
		if rw.Len() == 0 {
			t.Errorf("nothing written by enc.Encode(in)")
			return
		}
		dec := json.NewDecoder(rw)
		if err := dec.Decode(out); err != nil {
			t.Errorf("dec.Decode(out): %v", err)
			return
		}
		equals(t, in, out)
	})

	t.Run("YAML", func(t *testing.T) {
		out := &Qualifier{}
		rw := &bytes.Buffer{}
		enc := yaml.NewEncoder(rw)
		if err := enc.Encode(in); err != nil {
			t.Errorf("enc.Encode(in): %v", err)
			return
		}
		if rw.Len() == 0 {
			t.Errorf("nothing written by enc.Encode(in)")
			return
		}
		dec := yaml.NewDecoder(rw)
		if err := dec.Decode(out); err != nil {
			t.Errorf("dec.Decode(out): %v", err)
			return
		}
		equals(t, in, out)
	})

	t.Run("MsgPack", func(t *testing.T) {
		out := &Qualifier{}
		rw := &bytes.Buffer{}
		enc := msgpack.NewEncoder(rw)
		if err := enc.Encode(in); err != nil {
			t.Errorf("enc.Encode(in): %v", err)
			return
		}
		if rw.Len() == 0 {
			t.Errorf("nothing written by enc.Encode(in)")
			return
		}
		dec := msgpack.NewDecoder(rw)
		if err := dec.Decode(out); err != nil {
			t.Errorf("dec.Decode(out): %v", err)
			return
		}
		equals(t, in, out)
	})
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
