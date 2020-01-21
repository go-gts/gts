package gts

import (
	"bytes"
	"encoding/json"
	"testing"

	pars "gopkg.in/pars.v2"
	msgpack "gopkg.in/vmihailenco/msgpack.v4"
	yaml "gopkg.in/yaml.v3"
)

func TestGenBankIO(t *testing.T) {
	in := ReadGolden(t)
	state := pars.FromString(in)
	parser := pars.AsParser(GenBankParser)
	result, err := parser.Parse(state)
	if err != nil {
		t.Errorf("parser returned %v\nBuffer:\n%q", err, string(result.Token))
	}

	switch gb := result.Value.(type) {
	case *GenBank:
		t.Run("JSON", func(t *testing.T) {
			out := &GenBank{}
			rw := &bytes.Buffer{}
			enc := json.NewEncoder(rw)
			if err := enc.Encode(gb); err != nil {
				t.Errorf("enc.Encode(gb): %v", err)
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
			equals(t, gb.Fields, out.Fields)
			equals(t, gb.Features, out.Features)
			if !bytes.Equal(gb.Bytes(), out.Bytes()) {
				t.Error("gb.Bytes() != out.Bytes()")
				return
			}
		})

		t.Run("YAML", func(t *testing.T) {
			out := &GenBank{}
			rw := &bytes.Buffer{}
			enc := yaml.NewEncoder(rw)
			if err := enc.Encode(gb); err != nil {
				t.Errorf("enc.Encode(gb): %v", err)
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
			equals(t, gb.Fields, out.Fields)
			equals(t, gb.Features, out.Features)
			if !bytes.Equal(gb.Bytes(), out.Bytes()) {
				t.Error("gb.Bytes() != out.Bytes()")
				return
			}
		})

		t.Run("MsgPack", func(t *testing.T) {
			out := &GenBank{}
			rw := &bytes.Buffer{}
			enc := msgpack.NewEncoder(rw)
			if err := enc.Encode(gb); err != nil {
				t.Errorf("enc.Encode(gb): %v", err)
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
			equals(t, gb.Fields, out.Fields)
			equals(t, gb.Features, out.Features)
			if !bytes.Equal(gb.Bytes(), out.Bytes()) {
				t.Error("gb.Bytes() != out.Bytes()")
				return
			}
		})
	default:
		t.Errorf("result.Value.(type) = %T, want %T", gb, GenBank{})
	}
}
