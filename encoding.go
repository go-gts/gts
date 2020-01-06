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

// Decoder implement the Decode method for reading and decoding from an input
// stream to the object pointed by the given value.
type Decoder interface {
	Decode(v interface{}) error
}

// EncoderConstructor represents a constuctor function for a Decoder.
type DecoderConstructor func(io.Reader) Decoder

// Decodable implements the DecodeWith method whch should decode the object
// using the given Decoder object.
type Decodable interface {
	DecodeWith(dec Decoder) error
}

// NewJSONDecoder will create a JSON Decoder.
func NewJSONDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

// NewGobDecoder will create a Gob Decoder.
func NewGobDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}

// NewYAMLDecoder will create a YAML Decoder.
func NewYAMLDecoder(r io.Reader) Decoder {
	return yaml.NewDecoder(r)
}

// NewMsgpackDecoder will create a Msgpack Decoder.
func NewMsgpackDecoder(r io.Reader) Decoder {
	return msgpack.NewDecoder(r)
}

// EncodeJSON will encode the Encodable object as JSON.
func EncodeJSON(v Encodable) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := v.EncodeWith(enc)
	return buf.Bytes(), err
}

// DecodeJSON will decode the Decodable object from JSON.
func DecodeJSON(data []byte, v Decodable) error {
	buf := bytes.NewBuffer(data)
	dec := json.NewDecoder(buf)
	return v.DecodeWith(dec)
}

// EncodeGob will encode the Encodable object as Gob.
func EncodeGob(v Encodable) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := v.EncodeWith(enc)
	return buf.Bytes(), err
}

// DecodeGob will decode the Decodable object from Gob.
func DecodeGob(data []byte, v Decodable) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return v.DecodeWith(dec)
}
