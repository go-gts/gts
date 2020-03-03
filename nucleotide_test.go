package gts

import "testing"

func TestComplement(t *testing.T) {
	in := New(nil, []byte("ACGTURYKMBDHVacgturykmbdhv"))
	exp := New(nil, []byte("TGCAAYRMKVHDBtgcaayrmkvhdb"))
	out := Complement(in)
	equals(t, out, exp)
}

func TestTranscribe(t *testing.T) {
	in := New(nil, []byte("ACGTURYKMBDHVacgturykmbdhv"))
	exp := New(nil, []byte("UGCAAYRMKVHDBtgcaayrmkvhdb"))
	out := Transcribe(in)
	equals(t, out, exp)
}
