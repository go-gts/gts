package seqio

import "testing"

var detectTests = []struct {
	in  string
	out FileType
}{
	{"foo", DefaultFile},
	{"foo.fasta", FastaFile},
	{"foo.fastq", FastqFile},
	{"foo.gb", GenBankFile},
	{"foo.genbank", GenBankFile},
	{"foo.emb", EMBLFile},
	{"foo.embl", EMBLFile},
}

func TestDetect(t *testing.T) {
	for _, tt := range detectTests {
		out := Detect(tt.in)
		if out != tt.out {
			t.Errorf("Detect(%q) = %v, want %v", tt.in, out, tt.out)
		}
	}
}
