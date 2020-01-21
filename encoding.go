package gts

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"

	msgpack "gopkg.in/vmihailenco/msgpack.v4"
	yaml "gopkg.in/yaml.v3"
)

// Encoder implements the Encode method for encoding and writing the given
// object in some serialized form.
type Encoder interface {
	Encode(v interface{}) error
}

// EncoderConstructor represents a constuctor function for an Encoder.
type EncoderConstructor func(io.Writer) Encoder

// Encodable implements the EncodeTo method which should encode the object
// using the given Encoder object.
type Encodable interface {
	EncodeWith(enc Encoder) error
}

// NewJSONEncoder creates a JSON Encoder.
func NewJSONEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

// NewGobEncoder creates a Gob Encoder.
func NewGobEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}

// NewYAMLEncoder creates a YAML Encoder.
func NewYAMLEncoder(w io.Writer) Encoder {
	return yaml.NewEncoder(w)
}

// NewMsgpackEncoder creates a Msgpack Encoder.
func NewMsgpackEncoder(w io.Writer) Encoder {
	return msgpack.NewEncoder(w)
}

// EncoderWriter provides an interface for encoding objects with the given
// Encoder.
type EncoderWriter struct {
	value interface{}
	ctor  EncoderConstructor
}

// NewEncoderWriter creates a new EncoderWriter.
func NewEncoderWriter(v interface{}, ctor EncoderConstructor) EncoderWriter {
	return EncoderWriter{v, ctor}
}

// WriteTo satisfies the WriterTo interface.
func (f EncoderWriter) WriteTo(w io.Writer) (int64, error) {
	b := &bytes.Buffer{}
	enc := f.ctor(b)
	if err := enc.Encode(f.value); err != nil {
		return 0, err
	}
	n, err := w.Write(b.Bytes())
	return int64(n), err
}

// Decoder implement the Decode method for reading and decoding from an input
// stream to the object pointed by the given value.
type Decoder interface {
	Decode(v interface{}) error
}

// DecoderConstructor represents a constuctor function for a Decoder.
type DecoderConstructor func(io.Reader) Decoder

// Decodable implements the DecodeWith method whch should decode the object
// using the given Decoder object.
type Decodable interface {
	DecodeWith(dec Decoder) error
}

// NewJSONDecoder creates a JSON Decoder.
func NewJSONDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

// NewGobDecoder creates a Gob Decoder.
func NewGobDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}

// NewYAMLDecoder creates a YAML Decoder.
func NewYAMLDecoder(r io.Reader) Decoder {
	return yaml.NewDecoder(r)
}

// NewMsgpackDecoder creates a Msgpack Decoder.
func NewMsgpackDecoder(r io.Reader) Decoder {
	return msgpack.NewDecoder(r)
}

// EncodeJSON encodes the Encodable object as JSON.
func EncodeJSON(v Encodable) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := v.EncodeWith(enc)
	return buf.Bytes(), err
}

// DecodeJSON decodes the Decodable object from JSON.
func DecodeJSON(data []byte, v Decodable) error {
	buf := bytes.NewBuffer(data)
	dec := json.NewDecoder(buf)
	return v.DecodeWith(dec)
}

// EncodeGob encodes the Encodable object as Gob.
func EncodeGob(v Encodable) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := v.EncodeWith(enc)
	return buf.Bytes(), err
}

// DecodeGob decodes the Decodable object from Gob.
func DecodeGob(data []byte, v Decodable) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return v.DecodeWith(dec)
}
