package gts

import "testing"

type Stringer string

func (s Stringer) String() string { return string(s) }

func TestSeq(t *testing.T) {
	s := "atgc"
	p := []byte(s)

	seqs := Seq(s)
	seqp := Seq(p)
	seq := Seq(seqs)

	equals(t, seqs, seq)
	equals(t, seqp, seq)

	equals(t, Equal(seqs, seq), true)
	equals(t, Equal(seqp, seq), true)

	PanicTest(t, func() { Seq(0) })
}

func TestBareSequence(t *testing.T) {
	seq0 := Seq("atgc")
	seq1 := Seq("atgc")

	if err := seq0.Insert(2, seq1); err != nil {
		t.Errorf(
			"Seq(%q).Insert(2, Seq(%q)): %v",
			string(seq0.Data()),
			string(seq1.Data()),
			err,
		)
	}
	equals(t, seq0, Seq("atatgcgc"))
	if err := seq0.Replace(2, Complement(seq1)); err != nil {
		t.Errorf(
			"Seq(%q).Replace(2, Complement(Seq(%q))): %v",
			string(seq0.Data()),
			string(seq1.Data()),
			err,
		)
	}
	equals(t, seq0, Seq("attacggc"))
	if err := seq0.Delete(2, 4); err != nil {
		t.Errorf(
			"Seq(%q).Delete(2, 4): %v",
			string(seq0.Data()),
			err,
		)
	}
	equals(t, seq0, seq1)

	if seq0.Insert(4, seq1) == nil {
		t.Errorf(
			"Seq(%q).Insert(4, Seq(%q)) = nil, want error",
			string(seq0.Data()),
			string(seq1.Data()),
		)
	}
	equals(t, seq0, seq1)

	if seq0.Delete(4, 4) == nil {
		t.Errorf(
			"Seq(%q).Delete(4, 4) = nil, want error",
			string(seq0.Data()),
		)
	}
	equals(t, seq0, seq1)

	if seq0.Delete(1, 4) == nil {
		t.Errorf(
			"Seq(%q).Delete(1, 4) = nil, want error",
			string(seq0.Data()),
		)
	}
	equals(t, seq0, seq1)

	if seq0.Replace(4, seq1) == nil {
		t.Errorf(
			"Seq(%q).Replace(4, Seq(%q)) = nil, want error",
			string(seq0.Data()),
			string(seq1.Data()),
		)
	}
	equals(t, seq0, seq1)

	if seq0.Replace(1, seq1) == nil {
		t.Errorf(
			"Seq(%q).Replace(1, Seq(%q)) = nil, want error",
			string(seq0.Data()),
			string(seq1.Data()),
		)
	}
	equals(t, seq0, seq1)
}

func TestSlice(t *testing.T) {
	seq := Seq("atatgcgc")
	e := Seq("atgc")
	s := Slice(seq, 2, 6)

	equals(t, s, e)
}

func TestFragment(t *testing.T) {
	seq := Seq("atgcatgc")

	e44 := Seq("atgc")
	f44 := Fragment(seq, 4, 4)

	e24 := Seq("at")
	f24 := Fragment(seq, 2, 4)

	e42 := []Sequence{Seq("atgc"), Seq("gcat")}
	f42 := Fragment(seq, 4, 2)

	equals(t, f44[0], e44)
	equals(t, f44[1], e44)

	equals(t, f24[0], e24)
	equals(t, f24[1], e24)

	equals(t, f42[0], e42[0])
	equals(t, f42[1], e42[1])
}

func TestComposition(t *testing.T) {
	seq := Seq("atgcatgc")
	c := Composition(seq)
	e := map[byte]int{'a': 2, 't': 2, 'g': 2, 'c': 2}

	equals(t, c, e)
}

func TestSkew(t *testing.T) {
	seq := Seq("atgcatgc")

	values := []struct {
		nSet string
		pSet string
		skew float64
	}{
		{"g", "c", 0.0},
		{"a", "t", 0.0},
		{"g", "", -1.0},
		{"", "g", 1.0},
	}

	for _, value := range values {
		nSet, pSet, skew := value.nSet, value.pSet, value.skew
		out := Skew(seq, nSet, pSet)
		if out != skew {
			t.Errorf("Skew(%q, %q, %q) = %f, want %f", seq, nSet, pSet, out, skew)
		}
	}
}

func TestSequenceServer(t *testing.T) {
	seq := Seq("atgc")

	server := NewSequenceServer(Seq("atgc"))
	defer server.Close()

	proxy := server.Proxy()

	if err := server.Insert(2, seq); err != nil {
		t.Errorf(
			"Seq(%q).Insert(2, Seq(%q)): %v",
			string(server.Data()),
			string(seq.Data()),
			err,
		)
	}
	equals(t, server.Data(), Seq("atatgcgc").Data())
	equals(t, proxy.Data(), Seq("atatgcgc").Data())

	if err := server.Replace(2, Complement(seq)); err != nil {
		t.Errorf(
			"Seq(%q).Replace(2, Complement(Seq(%q))): %v",
			string(server.Data()),
			string(seq.Data()),
			err,
		)
	}
	equals(t, server.Data(), Seq("attacggc").Data())
	equals(t, proxy.Data(), Seq("attacggc").Data())

	if err := server.Delete(2, 4); err != nil {
		t.Errorf(
			"Seq(%q).Delete(2, 4): %v",
			string(server.Data()),
			err,
		)
	}
	equals(t, server.Data(), seq.Data())
	equals(t, proxy.Data(), seq.Data())

	if err := proxy.Insert(2, seq); err != nil {
		t.Errorf(
			"Seq(%q).Insert(2, Seq(%q)): %v",
			string(proxy.Data()),
			string(seq.Data()),
			err,
		)
	}
	equals(t, server.Data(), Seq("atatgcgc").Data())
	equals(t, proxy.Data(), Seq("atatgcgc").Data())

	if err := proxy.Replace(2, Complement(seq)); err != nil {
		t.Errorf(
			"Seq(%q).Replace(2, Complement(Seq(%q))): %v",
			string(proxy.Data()),
			string(seq.Data()),
			err,
		)
	}
	equals(t, server.Data(), Seq("attacggc").Data())
	equals(t, proxy.Data(), Seq("attacggc").Data())

	if err := proxy.Delete(2, 4); err != nil {
		t.Errorf(
			"Seq(%q).Delete(2, 4): %v",
			string(proxy.Data()),
			err,
		)
	}
	equals(t, server.Data(), seq.Data())
	equals(t, proxy.Data(), seq.Data())

	if proxy.Insert(4, seq) == nil {
		t.Errorf(
			"Seq(%q).Insert(4, Seq(%q)) = nil, want error",
			string(proxy.Data()),
			string(seq.Data()),
		)
	}
	equals(t, proxy.Data(), seq.Data())

	if proxy.Delete(4, 4) == nil {
		t.Errorf(
			"Seq(%q).Delete(4, 4) = nil, want error",
			string(proxy.Data()),
		)
	}
	equals(t, proxy.Data(), seq.Data())

	if proxy.Delete(1, 4) == nil {
		t.Errorf(
			"Seq(%q).Delete(1, 4) = nil, want error",
			string(proxy.Data()),
		)
	}
	equals(t, proxy.Data(), seq.Data())

	if proxy.Replace(4, seq) == nil {
		t.Errorf(
			"Seq(%q).Replace(4, Seq(%q)) = nil, want error",
			string(proxy.Data()),
			string(seq.Data()),
		)
	}
	equals(t, proxy.Data(), seq.Data())

	if proxy.Replace(1, seq) == nil {
		t.Errorf(
			"Seq(%q).Replace(1, Seq(%q)) = nil, want error",
			string(proxy.Data()),
			string(seq.Data()),
		)
	}
	equals(t, proxy.Data(), seq.Data())
}
