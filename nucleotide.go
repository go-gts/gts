package gts

import (
	"bytes"
	"regexp"
	"strings"
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
	ff := make([]Feature, len(seq.Features()))
	for i, f := range seq.Features() {
		ff[i].Key = f.Key
		ff[i].Location = f.Location.Complement()
		ff[i].Qualifiers = f.Qualifiers
	}
	return WithBytes(WithFeatures(seq, ff), p)
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

// Search for an oligomer within a sequence. The ambiguous nucleotides in the
// query sequence will match any of the respective nucleotides.
func Search(seq Sequence, query Sequence) []Location {
	if Len(seq) == 0 || Len(query) == 0 {
		return nil
	}

	b := strings.Builder{}
	for _, c := range query.Bytes() {
		switch c {
		case 't', 'u':
			b.WriteString("[tu]")
		case 'r':
			b.WriteString("[agr]")
		case 'y':
			b.WriteString("[ctuy]")
		case 'k':
			b.WriteString("[gtuy]")
		case 'm':
			b.WriteString("[acm]")
		case 's':
			b.WriteString("[cgs]")
		case 'w':
			b.WriteString("[atuw]")
		case 'b':
			b.WriteString("[cgtuyksb]")
		case 'd':
			b.WriteString("[agturkwd]")
		case 'h':
			b.WriteString("[actuymwh]")
		case 'v':
			b.WriteString("[acgrmsv]")
		case 'n':
			b.WriteString(".")
		default:
			b.WriteByte(c)
		}
	}

	s := b.String()
	p := bytes.ToLower(seq.Bytes())

	re := regexp.MustCompile(s)
	locs := re.FindAllIndex(p, -1)
	ret := make([]Location, len(locs))
	for i, loc := range locs {
		ret[i] = Range(loc[0], loc[1])
	}
	return ret
}
