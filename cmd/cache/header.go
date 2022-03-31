package cache

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

const HashSize = sha256.Size

// Header represents a cache header.
type Header []byte

// ReadHeader reads a header with the given hash from the reader.
func ReadHeader(r io.Reader) (Header, error) {
	p := make([]byte, HashSize*2)

	n, err := r.Read(p)
	if err != nil {
		return nil, fmt.Errorf("while reading header: %v", err)
	}

	if n != len(p) {
		return nil, errors.New("could not read sufficient bytes in header")
	}

	return p, nil
}

func (h Header) Block(n int) []byte {
	return h[n*HashSize : (n+1)*HashSize]
}

func (h Header) Root() []byte {
	return h[:HashSize]
}

func (h Header) Data() []byte {
	return h[HashSize:]
}

// Validate checks for the hash value integrity.
func (h Header) Validate(rsum, dsum, bsum []byte) error {
	if !bytes.Equal(rsum, h.Root()) {
		return errors.New("root hash sum mismatch")
	}

	if !bytes.Equal(dsum, h.Data()) {
		return errors.New("data hash sum mismatch")
	}

	return nil
}

// WriteTo writes the hash values to the given io.Writer.
func (h Header) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(h)
	return int64(n), err
}
