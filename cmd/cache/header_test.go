package cache

import (
	"bytes"
	"crypto/sha1"
	"testing"
)

func TestHeaderFail(t *testing.T) {
	h := sha1.New()
	b := &bytes.Buffer{}
	b.Write(h.Sum(nil))
	if _, err := ReadHeader(b, h.Size()); err == nil {
		t.Fatal("expected error in ReadHeader for insufficient read")
	}
}
