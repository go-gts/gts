package gts

import "testing"

func TestComplement(t *testing.T) {
	in := New(nil, nil, []byte("ACGTURYKMBDHVacgturykmbdhv"))
	exp := New(nil, nil, []byte("TGCAAYRMKVHDBtgcaayrmkvhdb"))
	out := Complement(in)
	equals(t, out, exp)
}

func TestTranscribe(t *testing.T) {
	in := New(nil, nil, []byte("ACGTURYKMBDHVacgturykmbdhv"))
	exp := New(nil, nil, []byte("UGCAAYRMKVHDBtgcaayrmkvhdb"))
	out := Transcribe(in)
	equals(t, out, exp)
}
