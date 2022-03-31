package seqio

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/internal/testutils"
	"github.com/go-pars/pars"
)

func TestFastaParser(t *testing.T) {
	files := []string{
		"NC_001422.fasta",
		"NC_001422_part.fasta",
		"sample.fasta",
	}

	for i, file := range files {
		testutils.RunCase(t, i, func(t *testing.T) {
			in := testutils.ReadTestfile(t, file)
			state := pars.FromString(in)
			parser := pars.AsParser(FastaParser)

			result, err := parser.Parse(state)
			if err != nil {
				t.Errorf("in file %q, parser returned %v\nBuffer:\n%s", file, err, string(result.Token))
			}

			if rec, ok := result.Value.(Record); !ok {
				t.Errorf("result.Value.(type) = %T, want %T", rec, Record{})
			}
		})
	}
}

var fastaParserFailTests = []string{
	// case 1
	"",

	// case 2
	"?",
}

func TestFastaParserFail(t *testing.T) {
	parser := pars.AsParser(FastaParser)
	for _, in := range fastaParserFailTests {
		state := pars.FromString(in)
		if err := parser(state, pars.Void); err == nil {
			t.Errorf("while parsing`\n%q\n`: expected error", in)
			return
		}
	}
}

func generateFASTATestRecords() []Record {
	length := 120
	records := make([]Record, 3)
	for i := range records {
		records[i].Header = fmt.Sprintf("sample%d", i)
		records[i].Sequence = gts.Seq(StringWithCharset("atgc", length))
	}
	return records
}

func TestFastaIOStream(t *testing.T) {
	buf := &bytes.Buffer{}
	state := pars.NewState(buf)

	stream := NewFastaIOStream(state, buf)
	fastaTestRecords := generateFASTATestRecords()

	for _, rec := range fastaTestRecords {
		s := rec.Header.(string)
		if err := stream.PushHeader(bytes.NewBufferString(s)); err != nil {
			t.Errorf("stream.PushHeader: %v", err)
		}
		if err := stream.PushFeatures(rec.Features); err != nil {
			t.Errorf("stream.PushFeatures: %v", err)
		}
		if err := stream.PushSequence(rec.Sequence); err != nil {
			t.Errorf("stream.PushSequence: %v", err)
		}
	}

	if stream.ForEach(func(i int, header interface{}, ff gts.Features) (SequenceHandler, error) {
		return nil, errors.New("error")
	}) == nil {
		t.Error("stream.ForEach: expected error")
	}

	if stream.ForEach(func(i int, header interface{}, ff gts.Features) (SequenceHandler, error) {
		return func(seq gts.Sequence) error {
			return errors.New("error")
		}, nil
	}) == nil {
		t.Error("stream.ForEach: expected error")
	}

	manip := func(i int, header interface{}, ff gts.Features) (SequenceHandler, error) {
		return func(seq gts.Sequence) error {
			testutils.RunCase(t, i, func(t *testing.T) {
				testutils.Equals(t, header, fastaTestRecords[i].Header)
				testutils.Equals(t, ff, fastaTestRecords[i].Features)
				out := seq.Bytes()
				exp := fastaTestRecords[i].Sequence.Bytes()
				fmt.Println(len(out), len(exp))
				testutils.RunCase(t, i, func(t *testing.T) {
					testutils.Diff(t, string(out), string(exp))
				})
			})
			return nil
		}, nil
	}

	if err := stream.ForEach(manip); err != nil {
		t.Errorf("stream.ForEach: %v", err)
	}

	if stream.PushHeader(nil) == nil {
		t.Errorf("stream.PushHeader: expected error")
	}

	if err := stream.PushFeatures(make(gts.Features, 1)); err != nil {
		t.Errorf("stream.PushFeatures: %v", err)
	}
}
