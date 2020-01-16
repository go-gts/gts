package gts

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	pars "gopkg.in/pars.v2"
	msgpack "gopkg.in/vmihailenco/msgpack.v4"
	yaml "gopkg.in/yaml.v3"
)

func TestFeatureIO(t *testing.T) {
	s := ReadGolden(t)
	ss := RecordSplit(s)

	for _, in := range ss {
		state := pars.FromString(in)
		parser := pars.Exact(FeatureListParser(""))
		result, err := parser.Parse(state)
		if err != nil {
			t.Errorf("while parsing`\n%s\n`: %v", in, err)
			return
		}
		switch ff := result.Value.(type) {
		case []Feature:
			f := ff[0]
			b := strings.Builder{}
			n, err := f.Format("     ", 21).WriteTo(&b)
			if err != nil {
				t.Errorf("f.WriteTo(w) = %d, %v, want %d, nil", n, err, n)
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
			t.Errorf("result.Value.(type) = %T, want %T", ff, Feature{})
		}
	}

	malformedKeyline := "     source          \n"

	for _, in := range []string{malformedKeyline} {
		state := pars.FromString(in)
		parser := pars.Exact(FeatureListParser(""))
		if _, err := parser.Parse(state); err == nil {
			t.Errorf("while parsing`\n%s\n`: expected error", in)
		}
	}
}

func TestFeatureEncoding(t *testing.T) {
	loc := NewRangeLocation(39, 723)
	qfs := Values{"foo": []string{"bar"}}
	in := &Feature{
		Key:        "misc_feature",
		Location:   loc,
		Qualifiers: qfs,
		order:      map[string]int{"foo": 0},
	}

	t.Run("JSON", func(t *testing.T) {
		out := &Feature{}
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
		out := &Feature{}
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
		out := &Feature{}
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

func TestFeatureListIO(t *testing.T) {
	in := ReadGolden(t)

	parser := pars.Exact(FeatureListParser(""))
	state := pars.FromString(in)
	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("while parsing`\n%s\n`: %v", in, err)
	}

	switch ff := result.Value.(type) {
	case []Feature:
		ft := FeatureList(ff)
		b := strings.Builder{}
		n, err := ft.Format("     ", 21).WriteTo(&b)
		if err != nil {
			t.Errorf("qf.WriteTo(w) = %d, %v, want %d, nil", n, err, n)
		}
		out := b.String() + "\n"
		equals(t, out, in)

		cp := FeatureList{}
		for _, f := range ft {
			cp.Add(f)
		}
		differs(t, cp, ft)
		sort.Sort(ByLocation(ft))
		equals(t, cp, ft)

		f := NewFeature("source", NewRangeLocation(39, 42), Values{})
		cp.Add(f)
		ft.Insert(len(ft)/2, f)
		differs(t, cp, ft)
		sort.Sort(ByLocation(ft))
		equals(t, cp, ft)
	default:
		t.Errorf("result.Value.(type) = %T, want %T", ff, []Feature{})
	}
}
