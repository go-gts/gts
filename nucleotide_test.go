package gts

import (
	"testing"

	"github.com/go-gts/gts/testutils"
)

func TestComplement(t *testing.T) {
	in := New(nil, nil, []byte("ACGTURYKMBDHVacgturykmbdhv."))
	exp := New(nil, nil, []byte("TGCAAYRMKVHDBtgcaayrmkvhdb."))
	out := Complement(in)
	testutils.Equals(t, out, exp)
}

func TestTranscribe(t *testing.T) {
	in := New(nil, nil, []byte("ACGTURYKMBDHVacgturykmbdhv."))
	exp := New(nil, nil, []byte("UGCAAYRMKVHDBtgcaayrmkvhdb."))
	out := Transcribe(in)
	testutils.Equals(t, out, exp)
}
