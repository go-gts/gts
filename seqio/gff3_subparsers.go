package seqio

/*
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

type gff3ParserContext struct {
	Header GFF3Header
	Seqs   map[string]*GFF3
	Order  []string
}

func newGFF3ParserContext() *gff3ParserContext {
	return &gff3ParserContext{
		GFF3Header{},
		make(map[string]*GFF3),
		make([]string, 0),
	}
}

func (ctx gff3ParserContext) Empty() bool {
	return len(ctx.Order) > 0
}

func (ctx *gff3ParserContext) Get(id string) *GFF3 {
	if seq, ok := ctx.Seqs[id]; ok {
		return seq
	}
	seq := &GFF3{}
	ctx.Order = append(ctx.Order, id)
	ctx.Seqs[id] = seq
	return seq
}

func (ctx *gff3ParserContext) Pop() *GFF3 {
	if ctx.Empty() {
		return nil
	}
	id, order := ctx.Order[0], ctx.Order[1:]
	seq, ok := ctx.Seqs[id]
	if !ok {
		return nil
	}
	ctx.Order = order
	delete(ctx.Seqs, id)

	// Copy common header values.
	seq.Header.Version = ctx.Header.Version
	seq.Header.AttributeOntology = ctx.Header.AttributeOntology
	seq.Header.FeatureOntology = ctx.Header.FeatureOntology
	seq.Header.SourceOntology = ctx.Header.SourceOntology

	// Find features with same ID.
	idmap := make(map[string][]gts.Feature)
	extra := []gts.Feature{}
	for _, f := range seq.Features() {
		if ids := f.Props.Get("ID"); len(ids) == 1 {
			idmap[ids[0]] = append(idmap[ids[0]], f)
		} else {
			extra = append(extra, f)
		}
	}

	// Join features with same ID.
	ff := gts.FeatureSlice{}
	for _, list := range idmap {
		locs := make([]gts.Location, len(list))
		for i, f := range list {
			locs[i] = f.Loc
		}
		ff.Insert(gts.NewFeature(
			list[0].Key,
			gts.Join(locs...),
			list[0].Props.Clone(),
		))
	}
	for _, f := range extra {
		ff.Insert(f)
	}

	seq.Table = ff

	return seq
}

var gff3VersionParser = pars.Seq("##gff-version ", pars.Line)

func gff3RegionParser(ctx *gff3ParserContext) pars.Parser {
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

		seq := ctx.Get(id)
		seq.Header.ID = id
		seq.Header.Region = gts.Segment{start, end}

		state.Drop()

		return nil
	}
}

func gff3OntologyParser(ctx *gff3ParserContext, ontologyType string) pars.Parser {
	prefix := pars.String(fmt.Sprintf("##%s-ontology ", ontologyType))
	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		if err := prefix(state, result); err != nil {
			state.Pop()
			return err
		}

		pars.Line(state, result)

		u, err := url.Parse(string(result.Token))
		if err != nil {
			pos := state.Position()
			state.Pop()
			return pars.NewError(err.Error(), pos)
		}

		switch ontologyType {
		case "feature":
			ctx.Header.FeatureOntology = u
		case "attribute":
			ctx.Header.AttributeOntology = u
		case "source":
			ctx.Header.SourceOntology = u
		}

		state.Drop()

		return nil
	}
}

func gff3SpeciesParser(ctx *gff3ParserContext) pars.Parser {
	prefix := pars.String("##species ")
	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		if err := prefix(state, result); err != nil {
			state.Pop()
			return err
		}

		pars.Line(state, result)

		u, err := url.Parse(string(result.Token))
		if err != nil {
			pos := state.Position()
			state.Pop()
			return pars.NewError(err.Error(), pos)
		}

		ctx.Header.Species = u

		state.Drop()

		return nil
	}
}

func gff3GenomeBuildParser(ctx *gff3ParserContext) pars.Parser {
	prefix := pars.String("##genome-build ")
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
		source := string(result.Token)

		// Guaranteed by Until(space).
		space(state, result)
		pars.Line(state, result)
		if len(result.Token) == 0 {
			pos := state.Position()
			state.Pop()
			return pars.NewError("expected build name", pos)
		}

		name := string(result.Token)
		ctx.Header.GenomeBuild = GFF3GenomeBuild{source, name}

		return nil
	}
}

func gff3TargetParser(target string) (string, gts.Location, bool) {
	index := -1
	if index = strings.Index(target, " "); index < 0 {
		return "", nil, false
	}

	seqid := target[:index]
	target = target[index+1:]

	if index = strings.Index(target, " "); index < 0 {
		return "", nil, false
	}
	start, err := strconv.Atoi(target[:index])
	if err != nil {
		return "", nil, false
	}
	target = target[index+1:]

	index = strings.Index(target, " ")
	if index < 0 {
		index = len(target)
	}
	end, err := strconv.Atoi(target[:index])
	if err != nil {
		return "", nil, false
	}
	target = target[index:]

	var loc gts.Location = gts.Range(start, end)
	switch len(target) {
	case 0:
		return seqid, loc, true
	case 2:
		switch target {
		case " +":
			return seqid, loc, true
		case " -":
			return seqid, loc.Complement(), true
		}
	}

	return "", nil, false
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

func gff3FastaParser(ctx *gff3ParserContext) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		scanner := NewScanner(FastaParser, state)
		for scanner.Scan() {
			fasta := scanner.Value().(Fasta)
			seq := ctx.Get(fasta.Desc)
			seq.FASTA = fasta
		}
		return scanner.Err()
	}
}
*/
