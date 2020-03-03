package gts

import "bytes"

// Complement returns the complement DNA sequence based on the FASTA sequence
// representation. All 'A's will be complemented to a 'T'. If the resulting
// sequence is intended to be RNA, use Transcribe instead.
func Complement(seq Sequence) Sequence {
	p := bytes.ReplaceAll(
		seq.Bytes(),
		[]byte("ACGTURYKMBDHVacgturykmbdhv"),
		[]byte("TGCAAYRMKVHDBtgcaayrmkvhdb"),
	)
	return New(seq.Info(), p)
}

// Transcribe returns the complement RNA sequence based on the FASTA sequence
// representation. All 'A's will be transcribed to a 'U'. If the resulting
// sequence is intended to be DNA, use Complement instead.
func Transcribe(seq Sequence) Sequence {
	p := bytes.ReplaceAll(
		seq.Bytes(),
		[]byte("ACGTURYKMBDHVacgturykmbdhv"),
		[]byte("UGCAAYRMKVHDBugcaayrmkvhdb"),
	)
	return New(seq.Info(), p)
}
