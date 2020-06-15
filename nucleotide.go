package gts

import (
	"bytes"
)

func replaceBytes(p, old, new []byte) []byte {
	q := make([]byte, len(p))
	for i, c := range p {
		switch j := bytes.IndexByte(old, c); j {
		case -1:
			q[i] = c
		default:
			q[i] = new[j]
		}
	}
	return q
}

// Complement returns the complement DNA sequence based on the FASTA sequence
// representation. All 'A's will be complemented to a 'T'. If the resulting
// sequence is intended to be RNA, use Transcribe instead.
func Complement(seq Sequence) Sequence {
	p := replaceBytes(
		seq.Bytes(),
		[]byte("ACGTURYKMBDHVacgturykmbdhv"),
		[]byte("TGCAAYRMKVHDBtgcaayrmkvhdb"),
	)
	return WithBytes(seq, p)
}

// Transcribe returns the complement RNA sequence based on the FASTA sequence
// representation. All 'A's will be transcribed to a 'U'. If the resulting
// sequence is intended to be DNA, use Complement instead.
func Transcribe(seq Sequence) Sequence {
	p := replaceBytes(
		seq.Bytes(),
		[]byte("ACGTURYKMBDHVacgturykmbdhv"),
		[]byte("UGCAAYRMKVHDBugcaayrmkvhdb"),
	)
	return WithBytes(seq, p)
}
