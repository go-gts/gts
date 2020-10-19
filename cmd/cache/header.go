package cache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// Header represents a cache header.
type Header struct {
	RootSum []byte
	DataSum []byte
	BodySum []byte
}

// ReadHeader reads a header with the given hash from the reader.
func ReadHeader(r io.Reader, size int) (Header, error) {
	p := make([]byte, size*3)

	n, err := r.Read(p)
	if err != nil {
		return Header{}, fmt.Errorf("while reading header: %v", err)
	}
	if n != len(p) {
		return Header{}, errors.New("could not read sufficient bytes in header")
	}

	i, j := size, size*2
	return Header{p[:i], p[i:j], p[j:]}, nil
}

// Validate checks for the hash value integrity.
func (h Header) Validate(rsum, dsum, bsum []byte) error {
	if !bytes.Equal(rsum, h.RootSum) {
		return errors.New("root hash sum mismatch")
	}
	if !bytes.Equal(dsum, h.DataSum) {
		return errors.New("data hash sum mismatch")
	}
	if !bytes.Equal(bsum, h.BodySum) {
		return errors.New("body hash sum mismatch")
	}
	return nil
}

// WriteTo writes the hash values to the given io.Writer.
func (h Header) WriteTo(w io.Writer) (int64, error) {
	p := append(append(h.RootSum, h.DataSum...), h.BodySum...)
	n, err := w.Write(p)
	return int64(n), err
}
