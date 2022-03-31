package seqio

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
	"github.com/go-wrap/wrap"
)

func fastaBodyParser(state *pars.State, result *pars.Result) error {
	b := bytes.Buffer{}

	for {
		state.Push()
		pars.Line(state, result)

		if bytes.IndexByte(result.Token, '>') == 0 {
			state.Pop()
			result.SetToken(b.Bytes())
			return nil
		}

		b.Write(result.Token)

		if state.Request(1) != nil {
			state.Clear()
			result.SetToken(b.Bytes())
			return nil
		}
	}
}

func FastaParser(state *pars.State, result *pars.Result) error {
	state.Push()

	c, err := state.ReadByte()
	if err != nil {
		return err
	}
	if c != '>' {
		state.Pop()
		return pars.NewError("expected `>`", state.Position())
	}

	pars.Line(state, result)
	header := string(result.Token)

	state.Clear()

	fastaBodyParser(state, result)

	result.SetValue(Record{header, nil, gts.AsSequence(result.Token)})

	return nil
}

type FastaIOStream struct {
	state  *pars.State
	result *pars.Result
	index  int

	w  io.Writer
	hh []string
	ss []gts.Sequence
}

func NewFastaIOStream(state *pars.State, w io.Writer) IOStream {
	return &FastaIOStream{state, &pars.Result{}, 0, w, nil, nil}
}

func (stream *FastaIOStream) Peek() error {
	stream.state.Push()
	if err := FastaParser(stream.state, stream.result); err != nil {
		stream.state.Pop()
		return err
	}
	stream.state.Drop()
	return nil
}

func (stream *FastaIOStream) Next(manip FeatureHandler) error {
	if stream.result.Value == nil {
		if err := stream.Peek(); err != nil {
			return err
		}
	}

	rec := stream.result.Value.(Record)
	if err := rec.Manipulate(manip, stream.index); err != nil {
		return err
	}

	stream.result.SetValue(nil)
	stream.index++

	return nil
}

func (stream *FastaIOStream) ForEach(manip FeatureHandler) error {
	for {
		if err := stream.Next(manip); err != nil {
			if dig(err) == io.EOF {
				return nil
			}
			return err
		}
	}
}

func (stream *FastaIOStream) PushHeader(header interface{}) error {
	switch v := header.(type) {
	case string:
		stream.hh = append(stream.hh, strings.ReplaceAll(v, "\n", " "))
		return stream.tryWrite()
	case fmt.Stringer:
		return stream.PushHeader(v.String())
	default:
		return fmt.Errorf("gts does not know how to format a sequence with header type `%T` as FASTA", v)
	}
}

func (stream *FastaIOStream) PushFeatures(ff gts.Features) error {
	if len(ff) > 0 {
		gts.Warnln("writing features is a no-op in FASTA files")
	}
	return nil
}

func (stream *FastaIOStream) PushSequence(seq gts.Sequence) error {
	stream.ss = append(stream.ss, seq)
	return stream.tryWrite()
}

func (stream *FastaIOStream) tryWrite() error {
	if len(stream.hh) == 0 || len(stream.ss) == 0 {
		return nil
	}

	desc, seq := stream.hh[0], stream.ss[0]
	stream.hh, stream.ss = stream.hh[1:], stream.ss[1:]

	data := wrap.Force(string(seq.Bytes()), 70)
	s := fmt.Sprintf(">%s\n%s\n", desc, data)
	_, err := io.WriteString(stream.w, s)
	return err
}
