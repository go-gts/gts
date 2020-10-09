package seqio

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-ascii/ascii"
	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
)

const spaceByte = byte(' ')

func joinByte(c byte) pars.Map {
	return pars.Join([]byte{c})
}

var errGenBankField = errors.New("expected field name")
var errGenBankExtra = errors.New("failed to parse unknown field")

func genbankFieldNameParser(q interface{}, depth int) pars.Parser {
	nameParser := pars.AsParser(q).Error(errGenBankField)
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
	return pars.Seq(indentParser, pars.Line).Child(1)
}

func genbankFieldBodyParser(depth int) pars.Parser {
	fieldLineParser := genbankFieldLineParser(depth)
	return func(state *pars.State, result *pars.Result) error {
		pars.Line(state, result)
		p := result.Token
		pars.Many(fieldLineParser)(state, result)
		children := make([]pars.Result, len(result.Children)+1)
		children[0] = *pars.NewTokenResult(p)
		for i, child := range result.Children {
			children[i+1] = child
		}
		result.SetChildren(children)
		return nil
	}
}

func genbankGenericFieldParser(name string, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser(name, depth)
	fieldBodyParser := genbankFieldBodyParser(depth).Map(joinByte('\n'))
	return pars.Seq(fieldNameParser, fieldBodyParser).Child(1)
}

func genbankExtraFieldParser(gb *GenBank, depth int) pars.Parser {
	fieldNamePattern := pars.Word(ascii.IsUpper)
	fieldNameParser := genbankFieldNameParser(fieldNamePattern, depth)
	fieldBodyparser := genbankFieldBodyParser(depth).Map(joinByte('\n'))
	fieldParser := pars.Seq(fieldNameParser, fieldBodyparser).Error(errGenBankExtra)
	return fieldParser.Map(func(result *pars.Result) error {
		name := string(result.Children[0].Token)
		value := string(result.Children[1].Token)
		extra := GenBankExtraField(name, value)
		gb.Fields.Extra = append(gb.Fields.Extra, extra)
		return nil
	})
}

func genbankSubfieldNameParser(name string, depth int) pars.Parser {
	nameParser := pars.String(name)
	paddingParser := pars.Many(spaceByte).Map(pars.Cat)
	parser := pars.Seq(paddingParser, nameParser, paddingParser)
	return func(state *pars.State, result *pars.Result) error {
		if err := parser(state, result); err != nil {
			return err
		}
		prefixLength := len(result.Children[0].Token)
		suffixLength := len(result.Children[2].Token)
		if prefixLength+len(name)+suffixLength != depth {
			what := fmt.Sprintf("uneven indent in subfield `%s`", name)
			return pars.NewError(what, state.Position())
		}
		return nil
	}
}

func genbankGenericSubfieldParser(name string, depth int) pars.Parser {
	subfieldNameParser := genbankSubfieldNameParser(name, depth)
	subfieldBodyParser := genbankFieldBodyParser(depth)
	return pars.Seq(subfieldNameParser, subfieldBodyParser).Child(1)
}

type genbankSubparser func(gb *GenBank, depth int) pars.Parser

func genbankDefinitionParser(gb *GenBank, depth int) pars.Parser {
	parser := genbankGenericFieldParser("DEFINITION", depth)
	return parser.Map(func(result *pars.Result) error {
		p := bytes.TrimSuffix(result.Token, []byte{'.'})
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

func genbankDBLinkParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("DBLINK", depth)
	pairParser := pars.Seq(
		pars.Until(':'), ": ", pars.Line,
	).Map(func(result *pars.Result) error {
		db := string(result.Children[0].Token)
		id := string(result.Children[2].Token)
		gb.Fields.DBLink.Set(db, id)
		return nil
	})
	indentParser := pars.String(strings.Repeat(" ", depth))
	extraParser := pars.Many(pars.Seq(indentParser, pairParser))
	return pars.Seq(fieldNameParser, pairParser, extraParser)
}

func genbankKeywordsParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("KEYWORDS", depth)
	fieldBodyParser := genbankFieldBodyParser(depth).Map(joinByte(' '))
	fieldParser := pars.Seq(fieldNameParser, fieldBodyParser).Child(1)
	return fieldParser.Map(func(result *pars.Result) error {
		gb.Fields.Keywords = FlatFileSplit(string(result.Token))
		return nil
	})
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

		taxonParser := pars.Many(fieldLineParser).Map(joinByte(spaceByte))
		taxonParser(state, result)
		gb.Fields.Source.Taxon = FlatFileSplit(string(result.Token))

		return nil
	}
}

func genbankReferenceSubfieldParser(ref *Reference, depth int) pars.Parser {
	subfieldParser := func(name string, depth int) pars.Parser {
		return genbankGenericSubfieldParser(name, depth).Map(joinByte('\n'))
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
	referenceLineParser := pars.Seq(fieldNameParser, pars.Int).Child(1)
	return func(state *pars.State, result *pars.Result) error {
		if err := referenceLineParser(state, result); err != nil {
			return err
		}

		ref := Reference{Number: result.Value.(int)}

		paddingLength := 3 - len(strconv.Itoa(ref.Number))
		paddingParser := pars.String(strings.Repeat(" ", paddingLength))
		infoParser := pars.Seq(pars.Maybe(paddingParser), pars.Line).Child(1)
		infoParser(state, result)
		ref.Info = string(result.Token)

		subfieldParser := genbankReferenceSubfieldParser(&ref, depth)
		pars.Many(subfieldParser)(state, result)

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
	fieldNameParser := pars.Seq("FEATURES", pars.Line)
	fieldBodyParser := gts.FeatureTableParser("")
	fieldParser := pars.Seq(fieldNameParser, pars.Cut, fieldBodyParser).Child(2)
	return fieldParser.Map(func(result *pars.Result) error {
		gb.Table = result.Value.(gts.FeatureTable)
		return nil
	})
}

func genbankContigParser(gb *GenBank, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("CONTIG", depth)
	accessionParser := pars.Seq(pars.Until(':'), ':').Child(0)
	locationParser := pars.Seq(pars.Int, "..", pars.Int).Children(0, 2)
	contigParser := pars.Seq("join(", accessionParser, locationParser, ')')
	fieldBodyParser := contigParser.Map(func(result *pars.Result) error {
		accession := string(result.Children[1].Token)
		head := result.Children[2].Children[0].Value.(int)
		tail := result.Children[2].Children[1].Value.(int)
		gb.Fields.Contig.Accession = accession
		gb.Fields.Contig.Region = gts.Segment{head - 1, tail}
		return nil
	})
	return pars.Seq(fieldNameParser, fieldBodyParser, pars.EOL)
}

func slowGenBankOriginParser(length int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		p := make([]byte, toOriginLength(length))
		offset := 0
		for i := 0; i < length; i += 60 {
			pars.Line(state, result)
			pattern := fmt.Sprintf("%9d", i+1)
			for j := 0; j < 60 && i+j < length; j += 10 {
				n := gts.Min(length-(i+j), 10)
				pattern += fmt.Sprintf(" \\w{%d}", n)
			}
			re := regexp.MustCompile(pattern)
			index := re.FindIndex(result.Token)
			if index == nil {
				return pars.NewError("malformed origin", state.Position())
			}
			head, tail := index[0], index[1]
			copy(p[offset:], result.Token[head:tail])
			offset += tail - head
			if offset < len(p) {
				p[offset] = '\n'
				offset++
			}
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

			state.Push()
			if err := state.Request(toOriginLength(length)); err != nil {
				return pars.NewError("not enough bytes in state", state.Position())
			}

			buffer := state.Buffer()
			state.Advance()

			if err := pars.EOL(state, result); err != nil {
				state.Pop()
				parser := slowGenBankOriginParser(length)
				if err := parser(state, result); err != nil {
					return err
				}
				buffer = result.Token
			} else {
				state.Drop()
			}
			gb.Origin = Origin{buffer, false}
			return nil
		}
	}
}
