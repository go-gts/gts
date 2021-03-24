package seqio

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-ascii/ascii"
	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
)

const spaceByte = byte(' ')

var isBaseCharacter = ascii.Range(33, 126)

var errGenBankField = errors.New("expected field name")
var errGenBankExtra = errors.New("failed to parse unknown field")

func genbankFieldNameParser(q interface{}, depth int) pars.Parser {
	nameParser := pars.AsParser(q)
	if s, ok := q.(string); ok {
		nameParser = pars.Bytes([]byte(s))
	}
	return func(state *pars.State, result *pars.Result) error {
		if err := nameParser(state, result); err != nil {
			return err
		}
		name := string(result.Token)
		indentLength := depth - len(name)
		indentParser := pars.String(strings.Repeat(" ", indentLength))
		paddingParser := pars.Any(indentParser, pars.Dry(pars.EOL))
		if paddingParser(state, pars.Void) != nil {
			state.Clear()
			what := fmt.Sprintf("uneven indent in field `%s`", name)
			return pars.NewError(what, state.Position())
		}
		return nil
	}
}

func genbankFieldLineParser(depth int) pars.Parser {
	indentParser := pars.String(strings.Repeat(" ", depth))
	return func(state *pars.State, result *pars.Result) error {
		if err := indentParser(state, pars.Void); err != nil {
			return pars.NewError("expected indent", state.Position())
		}
		return pars.Line(state, result)
	}
}

func genbankFieldBodyParser(depth int, sep byte) pars.Parser {
	fieldLineParser := genbankFieldLineParser(depth)
	return func(state *pars.State, result *pars.Result) error {
		pars.Line(state, result)
		w := bytes.NewBuffer(result.Token)
		for fieldLineParser(state, result) == nil {
			w.WriteByte(sep)
			w.Write(result.Token)
		}
		result.SetToken(w.Bytes())
		return nil
	}
}

func genbankGenericFieldParser(name string, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser(name, depth)
	fieldBodyParser := genbankFieldBodyParser(depth, '\n')
	return func(state *pars.State, result *pars.Result) error {
		if err := fieldNameParser(state, pars.Void); err != nil {
			return err
		}
		return fieldBodyParser(state, result)
	}
}

func genbankExtraFieldParser(gb *GenBank, depth int) pars.Parser {
	fieldNamePattern := pars.Word(ascii.IsUpper)
	fieldNameParser := genbankFieldNameParser(fieldNamePattern, depth)
	fieldBodyparser := genbankFieldBodyParser(depth, '\n')
	return func(state *pars.State, result *pars.Result) error {
		if fieldNameParser(state, result) != nil {
			return errGenBankExtra
		}
		name := string(result.Token)
		fieldBodyparser(state, result)
		value := string(result.Token)
		extra := GenBankExtraField(name, value)
		gb.Fields.Extra = append(gb.Fields.Extra, extra)
		return nil
	}
}

func genbankSubfieldNameParser(name string, depth int) pars.Parser {
	nameParser := pars.String(name)
	paddingParser := pars.Word(ascii.Is(spaceByte))
	return func(state *pars.State, result *pars.Result) error {
		paddingParser(state, result)
		prefixLength := len(result.Token)
		if prefixLength == 0 {
			return pars.NewError("expected indent", state.Position())
		}
		if err := nameParser(state, pars.Void); err != nil {
			return err
		}
		paddingParser(state, result)
		suffixLength := len(result.Token)
		if prefixLength+len(name)+suffixLength != depth {
			what := fmt.Sprintf("uneven indent in subfield `%s`", name)
			return pars.NewError(what, state.Position())
		}
		return nil
	}
}

func genbankGenericSubfieldParser(name string, depth int) pars.Parser {
	subfieldNameParser := genbankSubfieldNameParser(name, depth)
	subfieldBodyParser := genbankFieldBodyParser(depth, '\n')
	return func(state *pars.State, result *pars.Result) error {
		if err := subfieldNameParser(state, result); err != nil {
			return err
		}
		return subfieldBodyParser(state, result)
	}
}

type genbankSubparser func(gb *GenBank, depth int) pars.Parser

func genbankDefinitionParser(gb *GenBank, depth int) pars.Parser {
	parser := genbankGenericFieldParser("DEFINITION", depth)
	return parser.Map(func(result *pars.Result) error {
		p := result.Token
		if len(p) != 0 && p[len(p)-1] != '.' {
			return errors.New("expected period")
		}
		p = bytes.TrimSuffix(p, []byte{'.'})
		gb.Fields.Definition = string(p)
		return nil
	})
}

func genbankAccessionParser(gb *GenBank, depth int) pars.Parser {
	parser := genbankGenericFieldParser("ACCESSION", depth)
	return parser.Map(func(result *pars.Result) error {
		gb.Fields.Accession = string(result.Token)
		return nil
	})
}

func genbankVersionParser(gb *GenBank, depth int) pars.Parser {
	fieldParser := genbankGenericFieldParser("VERSION", depth)
	return fieldParser.Map(func(result *pars.Result) error {
		gb.Fields.Version = string(result.Token)
		return nil
	})
}

func genbankDBLinkPairParser(gb *GenBank, depth int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		pars.Line(state, result)
		s := string(result.Token)
		switch i := strings.IndexByte(s, ':'); i {
		case -1:
			return pars.NewError("expected `:`", state.Position())
		default:
			db, id := s[:i], s[i+2:]
			gb.Fields.DBLink.Set(db, id)
			return nil
		}
	}
}

func genbankDBLinkParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("DBLINK", depth)
	indentParser := pars.String(strings.Repeat(" ", depth))
	pairParser := genbankDBLinkPairParser(gb, depth)
	return func(state *pars.State, result *pars.Result) error {
		if err := fieldNameParser(state, pars.Void); err != nil {
			return err
		}
		if err := pairParser(state, result); err != nil {
			return err
		}
		for indentParser(state, pars.Void) == nil {
			if err := pairParser(state, result); err != nil {
				return err
			}
		}
		return nil
	}
}

func genbankKeywordsParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("KEYWORDS", depth)
	fieldBodyParser := genbankFieldBodyParser(depth, ' ')
	return func(state *pars.State, result *pars.Result) error {
		if err := fieldNameParser(state, pars.Void); err != nil {
			return err
		}
		fieldBodyParser(state, result)
		gb.Fields.Keywords = FlatFileSplit(string(result.Token))
		return nil
	}
}

func genbankSourceParser(gb *GenBank, depth int) pars.Parser {
	sourceParser := genbankGenericFieldParser("SOURCE", depth)
	sourceParser = sourceParser.Map(func(result *pars.Result) error {
		gb.Fields.Source.Species = string(result.Token)
		return nil
	})
	organismParser := genbankSubfieldNameParser("ORGANISM", depth)
	fieldLineParser := genbankFieldLineParser(depth)

	return func(state *pars.State, result *pars.Result) error {
		if err := sourceParser(state, result); err != nil {
			return err
		}

		if err := organismParser(state, pars.Void); err != nil {
			state.Pop()
			return err
		}

		pars.Line(state, result)
		gb.Fields.Source.Name = string(result.Token)

		w := bytes.Buffer{}
		for fieldLineParser(state, result) == nil {
			if w.Len() > 0 {
				w.WriteByte(spaceByte)
			}
			w.Write(result.Token)
		}
		gb.Fields.Source.Taxon = FlatFileSplit(string(w.Bytes()))

		return nil
	}
}

func genbankReferenceSubfieldParser(ref *Reference, depth int) pars.Parser {
	subfieldParser := func(name string, depth int) pars.Parser {
		return genbankGenericSubfieldParser(name, depth)
	}
	authorsParser := subfieldParser("AUTHORS", depth)
	consrtmParser := subfieldParser("CONSRTM", depth)
	titleParser := subfieldParser("TITLE", depth)
	journalParser := subfieldParser("JOURNAL", depth)
	pubmedParser := subfieldParser("PUBMED", depth)
	remarkParser := subfieldParser("REMARK", depth)
	return pars.Any(
		authorsParser.Map(func(result *pars.Result) error {
			ref.Authors = string(result.Token)
			return nil
		}),
		consrtmParser.Map(func(result *pars.Result) error {
			ref.Group = string(result.Token)
			return nil
		}),
		titleParser.Map(func(result *pars.Result) error {
			ref.Title = string(result.Token)
			return nil
		}),
		journalParser.Map(func(result *pars.Result) error {
			ref.Journal = string(result.Token)
			return nil
		}),
		pubmedParser.Map(func(result *pars.Result) error {
			ref.Xref = map[string]string{"PUBMED": string(result.Token)}
			return nil
		}),
		remarkParser.Map(func(result *pars.Result) error {
			ref.Comment = string(result.Token)
			return nil
		}),
	)
}

func genbankReferenceParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("REFERENCE", depth)
	return func(state *pars.State, result *pars.Result) error {
		if err := fieldNameParser(state, pars.Void); err != nil {
			return err
		}
		if err := pars.Int(state, result); err != nil {
			return err
		}

		ref := Reference{Number: result.Value.(int)}

		paddingLength := 3 - len(strconv.Itoa(ref.Number))
		paddingParser := pars.String(strings.Repeat(" ", paddingLength))
		paddingParser(state, pars.Void)
		pars.Line(state, result)
		ref.Info = string(result.Token)

		subfieldParser := genbankReferenceSubfieldParser(&ref, depth)
		for subfieldParser(state, result) == nil {
		}

		gb.Fields.References = append(gb.Fields.References, ref)

		return nil
	}
}

func genbankCommentParser(gb *GenBank, depth int) pars.Parser {
	fieldParser := genbankGenericFieldParser("COMMENT", depth)
	return fieldParser.Map(func(result *pars.Result) error {
		gb.Fields.Comments = append(gb.Fields.Comments, string(result.Token))
		return nil
	})
}

func genbankFeatureParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := pars.String("FEATURES")
	fieldBodyParser := INSDCTableParser("")
	return func(state *pars.State, result *pars.Result) error {
		if err := fieldNameParser(state, result); err != nil {
			return err
		}
		pars.Line(state, result)
		state.Clear()
		if err := fieldBodyParser(state, result); err != nil {
			return err
		}
		gb.Table = result.Value.([]gts.Feature)
		return nil
	}
}

func genbankContigParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("CONTIG", depth)
	untilColon := pars.Until(byte(':'))
	l, m, r := pars.String("join("), pars.String(".."), pars.Byte(')')
	return func(state *pars.State, result *pars.Result) error {
		if err := fieldNameParser(state, result); err != nil {
			return err
		}
		if err := l(state, pars.Void); err != nil {
			return err
		}
		if err := untilColon(state, result); err != nil {
			return err
		}
		accession := string(result.Token)
		pars.Skip(state, 1)
		if err := pars.Int(state, result); err != nil {
			return err
		}
		head := result.Value.(int)
		if err := m(state, pars.Void); err != nil {
			return err
		}
		if err := pars.Int(state, result); err != nil {
			return err
		}
		tail := result.Value.(int)
		if err := r(state, pars.Void); err != nil {
			return err
		}
		gb.Fields.Contig.Accession = accession
		gb.Fields.Contig.Region = gts.Segment{head - 1, tail}
		return nil
	}
}

func validateOrigin(p []byte, length int, pos pars.Position) error {
	offset := 0
	for i := 0; i < length; i += 60 {
		prefix := []byte(fmt.Sprintf("%9d", i+1))
		if !bytes.HasPrefix(p[offset:], prefix) {
			return pars.NewError("expected sequence index", pos)
		}
		offset += len(prefix)
		pos.Byte += len(prefix)

		for j := 0; j < 60 && i+j < length; j += 10 {
			if p[offset] != spaceByte {
				return pars.NewError("expected whitespace", pos)
			}
			offset++
			pos.Byte++

			for k := 0; k < 10 && i+j+k < length; k++ {
				if !isBaseCharacter(p[offset]) {
					return pars.NewError("expected character", pos)
				}
				offset++
				pos.Byte++
			}
		}

		if p[offset] != '\n' {
			return pars.NewError("expected newline", pos)
		}
		offset++

		pos.Line++
		pos.Byte = 0
	}

	return nil
}

func slowGenBankOriginParser(length int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		p := make([]byte, toOriginLength(length))
		offset := 0
		for i := 0; i < length; i += 60 {
			pos := state.Position()
			pars.Line(state, result)

			q := result.Token
			extent := 0
			prefix := []byte(fmt.Sprintf("%9d", i+1))
			if !bytes.HasPrefix(q, prefix) {
				return pars.NewError("expected sequence index", pos)
			}
			extent += len(prefix)

			for j := 0; j < 60 && i+j < length; j += 10 {
				if q[extent] != spaceByte {
					pos.Byte += extent
					return pars.NewError("expected whitespace", pos)
				}
				extent++

				for k := 0; k < 10 && i+j+k < length; k++ {
					if !isBaseCharacter(q[extent]) {
						pos.Byte += extent
						return pars.NewError("expected character", pos)
					}
					extent++
				}
			}

			offset += copy(p[offset:], q[:extent])
			p[offset] = '\n'
			offset++
		}
		result.SetToken(p)
		return nil
	}
}

func makeGenbankOriginParser(length int) genbankSubparser {
	return func(gb *GenBank, depth int) pars.Parser {
		fieldNameParser := genbankFieldNameParser("ORIGIN", depth)
		return func(state *pars.State, result *pars.Result) error {
			if err := fieldNameParser(state, result); err != nil {
				return err
			}
			pars.Line(state, result)

			if err := state.Request(toOriginLength(length)); err != nil {
				return pars.NewError("not enough bytes in state", state.Position())
			}

			p := state.Buffer()
			if validateOrigin(p, length, state.Position()) == nil {
				state.Advance()
				gb.Origin = &Origin{p, false}
				return nil
			}

			parser := slowGenBankOriginParser(length)
			if err := parser(state, result); err != nil {
				return err
			}
			p = result.Token

			gb.Origin = &Origin{p, false}
			return nil
		}
	}
}
