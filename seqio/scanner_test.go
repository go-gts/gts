package seqio

import (
	"testing"

	"github.com/go-gts/gts/testutils"
	"github.com/go-pars/pars"
)

func TestScannerGenBank(t *testing.T) {
	in := testutils.ReadTestfile(t, "NC_001422.gb")
	s := NewScanner(GenBankParser, pars.FromString(in))
	if !s.Scan() {
		if s.Err() == nil {
			t.Error("Scan failed but returned nil error")
			return
		}
	}

	if s.Err() != nil {
		t.Errorf("Scan failed: %v", s.Err())
		return
	}

	if seq, ok := s.Value().(GenBank); !ok {
		t.Errorf("result.Value.(type) = %T, want %T", seq, GenBank{})
	}
}

func TestScannerGenBankFail(t *testing.T) {
	in := testutils.ReadTestfile(t, "NC_001422.fasta")
	s := NewScanner(GenBankParser, pars.FromString(in))
	if s.Scan() {
		t.Error("GenBank Scanner should fail for FASTA file")
		return
	}
	if s.Err() == nil {
		t.Error("expected error in GenBank Scanner")
		return
	}
	if s.Scan() {
		t.Error("Scanner should halt after first error")
		return
	}
}

func TestAutoScanner(t *testing.T) {
	in := testutils.ReadTestfile(t, "NC_001422.fasta")
	s := NewAutoScanner(pars.FromString(in))
	if !s.Scan() {
		if s.Err() == nil {
			t.Error("Scan failed but returned nil error")
			return
		}
	}

	if s.Err() != nil {
		t.Errorf("Scan failed: %v", s.Err())
		return
	}

	if seq, ok := s.Value().(Fasta); !ok {
		t.Errorf("result.Value.(type) = %T, want %T", seq, GenBank{})
	}
}

func TestAutoScannerFail(t *testing.T) {
	in := "LOCUS       NC_001422               5386 bp ss-DNA     circular PHG 06-JUL-2018"
	s := NewAutoScanner(pars.FromString(in))
	if s.Scan() {
		t.Error("Auto Scanner should fail")
		return
	}
	if s.Err() == nil {
		t.Error("expected error in Auto Scanner")
		return
	}
}
