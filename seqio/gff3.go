package seqio

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
)

// GFF3GenomeBuild represents the genome build directive.
type GFF3GenomeBuild struct {
	Source string
	Name   string
}

// GFF3Header represents the directives of a GFF3 record other than the
// features and sequence.
type GFF3Header struct {
	Version string
	ID      string
	Region  gts.Segment

	FeatureOntology   *url.URL
	AttributeOntology *url.URL
	SourceOntology    *url.URL

	Species     *url.URL
	GenomeBuild GFF3GenomeBuild
}

func gff3RegionParser(headers map[string]GFF3Header) pars.Parser {
	prefix := pars.String("##sequence-region ")
	space := pars.Byte(' ')
	word := pars.Until(space)

	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		if err := prefix(state, result); err != nil {
			state.Pop()
			return err
		}
		if err := word(state, result); err != nil {
			state.Pop()
			return err
		}
		id := string(result.Token)

		// Guaranteed by Until(space).
		space(state, result)
		if err := pars.Int(state, result); err != nil {
			state.Pop()
			return err
		}
		start := result.Value.(int)

		if err := space(state, result); err != nil {
			state.Pop()
			return err
		}
		if err := pars.Int(state, result); err != nil {
			state.Pop()
			return err
		}
		end := result.Value.(int)

		headers[id] = GFF3Header{
			ID:     id,
			Region: gts.Segment{start, end},
		}

		state.Drop()

		return nil
	}
}

func gff3FeatureParser(ctx *gff3ParserContext) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		pos := state.Position()
		pars.Line(state, result)

		// Parse feature sequence ID.
		line, index := result.Token, 0
		if index = bytes.IndexByte(line, '\t'); index < 0 {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		seqid := string(line[:index])
		line = line[index+1:]
		pos.Byte += index + 1

		// Parse feature source.
		if index = bytes.IndexByte(line, '\t'); index < 0 {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		props := gts.Props{}
		props.Add("source", string(line[:index]))
		line = line[index+1:]
		pos.Byte += index + 1

		// Parse feature key.
		if index = bytes.IndexByte(line, '\t'); index < 0 {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		key := string(line[:index])
		line = line[index+1:]
		pos.Byte += index + 1

		// Parse feature start position.
		if index = bytes.IndexByte(line, '\t'); index < 0 {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		start, err := strconv.Atoi(string(line[:index]))
		if err != nil {
			state.Pop()
			return pars.NewError(err.Error(), pos)
		}
		line = line[index+1:]
		pos.Byte += index + 1

		// Parse feature end position.
		if index = bytes.IndexByte(line, '\t'); index < 0 {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		end, err := strconv.Atoi(string(line[:index]))
		if err != nil {
			state.Pop()
			return pars.NewError(err.Error(), pos)
		}
		line = line[index+1:]
		pos.Byte += index + 1

		loc := gts.Range(start, end)

		// Parse feature score.
		if index = bytes.IndexByte(line, '\t'); index < 0 {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		if _, err := strconv.ParseFloat(string(line[:index]), 64); err != nil {
			if string(line[:index]) != "." {
				state.Pop()
				return pars.NewError(err.Error(), pos)
			}
		}
		props.Add("score", string(line[:index]))
		line = line[index+1:]
		pos.Byte += index + 1

		// Parse feature strand.
		if len(line) == 0 {
			return pars.NewError("expected strand", pos)
		}
		switch strand := line[0]; strand {
		case '+', '-', '.', '?':
			props.Add("strand", string(strand))
		default:
			state.Pop()
			what := fmt.Sprintf(
				"unexpected strand value %s",
				strconv.QuoteRuneToGraphic(rune(strand)),
			)
			return pars.NewError(what, pos)
		}
		line = line[1:]
		pos.Byte++
		if len(line) == 0 || line[0] != '\t' {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		line = line[1:]
		pos.Byte++

		// Parse feature phase.
		if len(line) == 0 {
			return pars.NewError("expected phase", pos)
		}
		switch phase := line[0]; phase {
		case '.', '0', '1', '2':
			if key == "CDS" && phase == '.' {
				return pars.NewError("CDS requires a phase", pos)
			}
			props.Add("phase", string(phase))
		default:
			state.Pop()
			what := fmt.Sprintf(
				"unexpected phase value %s",
				strconv.QuoteRuneToGraphic(rune(phase)),
			)
			return pars.NewError(what, pos)
		}
		line = line[1:]
		pos.Byte++
		if len(line) == 0 || line[0] != '\t' {
			state.Pop()
			return pars.NewError("expected `\\t`", pos)
		}
		line = line[1:]
		pos.Byte++

		// Parse feature attributes.
		fields := strings.Split(string(line), ";")
		for _, field := range fields {
			switch index := strings.IndexByte(field, '='); index {
			case -1:
				state.Pop()
				return pars.NewError("expected `=`", pos)
			default:
				name, value := field[:index], field[index+1:]
				props.Add(name, value)
			}
			pos.Byte += len(field)
		}

		// Insert feature to sequence.
		seq := ctx.Get(seqid)
		seq.Table = append(seq.Table, gts.NewFeature(key, loc, props))

		// Insert feature to target sequence.
		for _, target := range props.Get("Target") {
			seqid, loc, ok := gff3TargetParser(target)
			if !ok {
				state.Drop()
				return pars.NewError("malformed Target attribute", pos)
			}
			seq := ctx.Get(seqid)
			seq.Table = append(seq.Table, gts.NewFeature(key, loc, props.Clone()))
		}

		state.Drop()

		return nil
	}
}

var gff3FastaDirective = pars.Seq("##FASTA", pars.EOL)

var errGFF3FastaDirective = errors.New("gff3FastaDirective")

func gff3FastaDirectiveParser(state *pars.State, result *pars.Result) error {
	if gff3FastaDirective(state, result) == nil {
		return errGFF3FastaDirective
	}
	return nil
}

type GFF3IStream struct {
	state  *pars.State
	result *pars.Result
	index  int
	peeked bool

	header GFF3Header
	tables map[string]gts.Features
}

func (istream *GFF3IStream) Peek() error {
	return nil
}

func (istream *GFF3IStream) ForEach(fh FeatureHandler) error {
	return nil
}
