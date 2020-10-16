package main

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
)

func newHash() hash.Hash {
	return sha1.New()
}

func encodeToString(p []byte) string {
	return hex.EncodeToString(p)
}
