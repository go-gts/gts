package gts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-ascii/ascii"
	"github.com/go-pars/pars"
	"github.com/go-wrap/wrap"
)

// GenBankFields represents the fields of a GenBank record other than the
// features and sequence.
type GenBankFields struct {
	LocusName string
	Molecule  string
	Topology  string
	Division  string
	Date      Date

	Definition string
	Accession  string
	Version    string
	DBLink     Dictionary
	Keywords   []string
	Source     Organism
	References []Reference
	Comment    string
}

// GenBank represents a GenBank sequence record.
type GenBank struct {
	Fields GenBankFields
	Table  FeatureTable
	Origin []byte
}

// Info returns the metadata of the sequence.
func (gb GenBank) Info() interface{} {
	return gb.Fields
}

// Features returns the feature table of the sequence.
func (gb GenBank) Features() FeatureTable {
	return gb.Table
}

type originIterator struct {
	line, block int
}

func (iter *originIterator) Next() (int, int) {
	i := iter.line*76 + iter.block*11 + 10
	j := iter.line*60 + iter.block*10
	iter.block++
	if iter.block == 6 {
		iter.block = 0
		iter.line++
	}
	return i, j
}

// Len returns the length of the sequence.
func (gb GenBank) Len() int {
	iter := originIterator{0, 0}
	i, _ := iter.Next()
	length := 0
	for i < len(gb.Origin) {
		n := min(10, len(gb.Origin)-i)
		length += n
		i, _ = iter.Next()
	}
	return length
}

// Bytes returns the byte representation of the sequence.
func (gb GenBank) Bytes() []byte {
	data := make([]byte, gb.Len())
	iter := originIterator{0, 0}
	i, j := iter.Next()
	for i < len(gb.Origin) {
		n := min(10, len(gb.Origin)-i)
		copy(data[j:j+n], gb.Origin[i:i+n])
		i, j = iter.Next()
	}
	return data
}

// String satisifes the fmt.Stringer interface.
func (gb GenBank) String() string {
	builder := strings.Builder{}
	indent := "            "

	date := strings.ToUpper(gb.Fields.Date.ToTime().Format("02-Jan-2006"))
	locus := fmt.Sprintf(
		"%-12s%-17s %10d bp %6s     %-9s%s %s", "LOCUS", gb.Fields.LocusName,
		gb.Len(), gb.Fields.Molecule, gb.Fields.Topology, gb.Fields.Division, date,
	)

	builder.WriteString(locus)

	definition := AddPrefix(gb.Fields.Definition, indent)
	builder.WriteString("\nDEFINITION  " + definition)
	builder.WriteString("\nACCESSION   " + gb.Fields.Accession)
	builder.WriteString("\nVERSION     " + gb.Fields.Version)

	for i, pair := range gb.Fields.DBLink {
		switch i {
		case 0:
			builder.WriteString("\nDBLINK      ")
		default:
			builder.WriteString("\n" + indent)
		}
		builder.WriteString(fmt.Sprintf("%s: %s", pair.Key, pair.Value))
	}

	keywords := wrap.Space(strings.Join(gb.Fields.Keywords, "; ")+".", 67)
	keywords = AddPrefix(keywords, indent)
	builder.WriteString("\nKEYWORDS    " + keywords)

	source := wrap.Space(gb.Fields.Source.Species, 67)
	source = AddPrefix(source, indent)
	builder.WriteString("\nSOURCE      " + source)

	organism := wrap.Space(gb.Fields.Source.Name, 67)
	organism = AddPrefix(organism, indent)
	builder.WriteString("\n  ORGANISM  " + organism)

	taxon := wrap.Space(strings.Join(gb.Fields.Source.Taxon, "; ")+".", 67)
	taxon = AddPrefix(taxon, indent)
	builder.WriteString("\n" + indent + taxon)

	for _, ref := range gb.Fields.References {
		builder.WriteString(fmt.Sprintf("\nREFERENCE   %-2d ", ref.Number))
		switch len(ref.Ranges) {
		case 0:
			builder.WriteString("(sites)")
		default:
			ranges := make([]string, len(ref.Ranges))
			for i, rng := range ref.Ranges {
				ranges[i] = fmt.Sprintf("%d to %d", rng.Start, rng.End)
			}
			builder.WriteString(fmt.Sprintf("(bases %s)", strings.Join(ranges, "; ")))
		}
		if ref.Authors != "" {
			builder.WriteString("\n  AUTHORS   " + AddPrefix(ref.Authors, indent))
		}
		if ref.Group != "" {
			builder.WriteString("\n  CONSRTM   " + AddPrefix(ref.Group, indent))
		}
		if ref.Title != "" {
			builder.WriteString("\n  TITLE     " + AddPrefix(ref.Title, indent))
		}
		if ref.Journal != "" {
			builder.WriteString("\n  JOURNAL   " + AddPrefix(ref.Journal, indent))
		}
		if ref.Xref != nil {
			if v, ok := ref.Xref["PUBMED"]; ok {
				builder.WriteString("\n   PUBMED   " + v)
			}
		}
		if ref.Comment != "" {
			builder.WriteString("\n  REMARK    " + AddPrefix(ref.Comment, indent))
		}
	}

	if gb.Fields.Comment != "" {
		builder.WriteString("\nCOMMENT     " + AddPrefix(gb.Fields.Comment, indent))
	}

	builder.WriteString("\nFEATURES             Location/Qualifiers\n")

	gb.Table.Format("     ", 21).WriteTo(&builder)

	builder.WriteString("\nORIGIN      \n")
	builder.Write(gb.Origin)
	builder.WriteString("\n//\n")

	return builder.String()
}

// WriteTo satisfies the io.WriterTo interface.
func (gb GenBank) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, gb.String())
	return int64(n), err
}

// GenBankFormatter attempts to format a record in GenBank flatfile format.
type GenBankFormatter struct {
	seq Sequence
}

// WriteTo satisfies the io.WriterTo interface.
func (gf GenBankFormatter) WriteTo(w io.Writer) (int64, error) {
	switch seq := gf.seq.(type) {
	case GenBank:
		return seq.WriteTo(w)
	case *GenBank:
		return GenBankFormatter{*seq}.WriteTo(w)
	default:
		return 0, fmt.Errorf("gts does not know how to format `%T` in GenBank flatfile form", seq)
	}
}

var genbankLocusParser = pars.Seq(
	"LOCUS", pars.Spaces,
	pars.Word(ascii.IsSnake), pars.Spaces,
	pars.Int, " bp", pars.Spaces,
	pars.Word(ascii.Not(ascii.IsSpace)), pars.Spaces,
	pars.Any("linear", "circular"), pars.Spaces,
	pars.Count(pars.Byte(), 3).Map(pars.Cat), pars.Spaces,
	pars.AsParser(pars.Line).Map(pars.Time("02-Jan-2006")),
).Children(1, 2, 4, 7, 9, 11, 13)

func genbankFieldBodyParser(depth int) pars.Parser {
	indent := pars.String(strings.Repeat(" ", depth))
	return func(state *pars.State, result *pars.Result) error {
		pars.Line(state, result)
		tmp := *pars.NewTokenResult(result.Token)
		parser := pars.Many(pars.Seq(indent, pars.Line).Child(1))
		parser(state, result)
		children := append([]pars.Result{tmp}, result.Children...)
		result.SetChildren(children)
		return nil
	}
}

// GenBankParser attempts to parse a single GenBank record.
func GenBankParser(state *pars.State, result *pars.Result) error {
	if err := genbankLocusParser(state, result); err != nil {
		return err
	}

	pars.Cut(state, result)

	depth := len(result.Children[0].Token) + 5
	indent := pars.String(strings.Repeat(" ", depth))

	locus := string(result.Children[1].Token)
	length := result.Children[2].Value.(int)
	molecule := string(result.Children[3].Token)
	topology := result.Children[4].Value.(string)
	division := string(result.Children[5].Token)
	date := FromTime(result.Children[6].Value.(time.Time))

	fields := GenBankFields{
		LocusName: locus,
		Molecule:  molecule,
		Topology:  topology,
		Division:  division,
		Date:      date,
	}

	gb := GenBank{Fields: fields}

	fieldNameParser := pars.Word(ascii.IsUpper).Error(errors.New("expected field name"))
	fieldBodyParser := genbankFieldBodyParser(depth)
	end := pars.Seq("//", pars.EOL).Error(errors.New("expected end of record"))

	for {
		if end(state, result) == nil {
			result.SetValue(gb)
			return nil
		}
		if err := fieldNameParser(state, result); err != nil {
			return err
		}
		name := string(result.Token)
		paddingParser := pars.Count(' ', depth-len(name))

		if err := paddingParser(state, result); name != "ORIGIN" && err != nil {
			return pars.NewError("uneven indent", state.Position())
		}

		switch name {
		case "DEFINITION":
			parser := fieldBodyParser.Map(pars.Join([]byte("\n")))
			parser(state, result)
			gb.Fields.Definition = string(result.Token)

		case "ACCESSION":
			pars.Line(state, result)
			gb.Fields.Accession = string(result.Token)

		case "VERSION":
			pars.Line(state, result)
			gb.Fields.Version = string(result.Token)

		case "DBLINK":
			headParser := pars.Seq(pars.Until(':'), ':', pars.Line).Children(0, 2)
			if err := headParser(state, result); err != nil {
				return err
			}
			db := string(result.Children[0].Token)
			id := string(result.Children[1].Token[1:])
			gb.Fields.DBLink.Set(db, id)

			tailParser := pars.Many(pars.Seq(indent, headParser).Child(1))
			tailParser(state, result)
			for _, child := range result.Children {
				db = string(child.Children[0].Token)
				id = string(child.Children[1].Token[1:])
				gb.Fields.DBLink.Set(db, id)
			}

		case "KEYWORDS":
			parser := fieldBodyParser.Map(pars.Join([]byte(" ")))
			parser(state, result)
			gb.Fields.Keywords = FlatFileSplit(string(result.Token))

		case "SOURCE":
			sourceParser := fieldBodyParser.Map(pars.Join([]byte("\n")))
			sourceParser(state, result)

			organism := Organism{}
			organism.Species = string(result.Token)

			organismLineParser := pars.Seq(
				pars.Spaces, []byte("ORGANISM"), pars.Spaces,
			).Map(pars.Cat)

			if organismLineParser(state, result) == nil {
				if len(result.Token) != depth {
					return pars.NewError("uneven indent", state.Position())
				}
				pars.Line(state, result)
				organism.Name = string(result.Token)

				taxonParser := pars.Many(
					pars.Seq(indent, pars.Line).Child(1),
				).Map(pars.Join([]byte(" ")))
				taxonParser(state, result)
				organism.Taxon = FlatFileSplit(string(result.Token))
			}

			gb.Fields.Source = organism

		case "REFERENCE":
			parser := fieldBodyParser.Map(pars.Join([]byte(" ")))
			parser(state, result)

			rangeParser := pars.Seq(
				pars.Int, pars.Spaces, '(',
				pars.Any(
					pars.Seq(
						pars.Any("bases ", "residues "),
						pars.Delim(pars.Seq(pars.Int, " to ", pars.Int).Children(0, 2), "; "),
					).Child(1),
					"sites",
				), ')', pars.EOL,
			).Children(0, 3)
			if err := rangeParser(pars.FromBytes(result.Token), result); err != nil {
				return pars.NewError("failed to parse reference", state.Position())
			}

			reference := Reference{}
			reference.Number = result.Children[0].Value.(int)

			if _, ok := result.Children[1].Value.(string); !ok {
				children := result.Children[1].Children
				ranges := make([]ReferenceRange, len(children))
				for i, child := range children {
					start := child.Children[0].Value.(int)
					end := child.Children[1].Value.(int)
					ranges[i] = ReferenceRange{start, end}
				}
				reference.Ranges = ranges
			}

			subfieldParser := pars.Seq(
				pars.Spaces,
				pars.Any(
					"AUTHORS",
					"CONSRTM",
					"TITLE",
					"JOURNAL",
					"PUBMED",
					"REMARK",
				),
				pars.Spaces,
			).Map(func(result *pars.Result) error {
				children := result.Children
				name := children[1].Value.(string)
				depth := len(name) + len(children[0].Token) + len(children[2].Token)
				*result = *pars.AsResults(name, depth)
				return nil
			})

			for subfieldParser(state, result) == nil {
				name := result.Children[0].Value.(string)
				if result.Children[1].Value.(int) != depth {
					return pars.NewError("uneven indent", state.Position())
				}
				parser := fieldBodyParser.Map(pars.Join([]byte("\n")))
				parser(state, result)
				switch name {
				case "AUTHORS":
					reference.Authors = string(result.Token)
				case "CONSRTM":
					reference.Group = string(result.Token)
				case "TITLE":
					reference.Title = string(result.Token)
				case "JOURNAL":
					reference.Journal = string(result.Token)
				case "PUBMED":
					reference.Xref = map[string]string{"PUBMED": string(result.Token)}
				case "REMARK":
					reference.Comment = string(result.Token)
				}
			}

			gb.Fields.References = append(gb.Fields.References, reference)

		case "COMMENT":
			parser := fieldBodyParser.Map(pars.Join([]byte("\n")))
			parser(state, result)
			gb.Fields.Comment = string(result.Token)

		case "FEATURES":
			pars.Line(state, result)
			parser := FeatureTableParser("")
			if err := parser(state, result); err != nil {
				return err
			}
			gb.Table = FeatureTable(result.Value.(FeatureTable))

		case "ORIGIN":
			// Trim off excess whitespace.
			pars.Line(state, result)

			// Check for origin syntax but do not parse.
			state.Push()
			for i := 0; i < length; i += 60 {
				prefix := []byte(fmt.Sprintf("% 9d", i+1))
				if err := state.Request(9); err != nil {
					return pars.NewError("not enough bytes in state", state.Position())
				}
				if !bytes.Equal(state.Buffer(), prefix) {
					return pars.NewError("malformed origin prefix", state.Position())
				}
				state.Advance()

				j := 0
				c, err := pars.Next(state)
				for err == nil && c != '\n' {
					if j%11 == 0 {
						if c != ' ' {
							return pars.NewError("expected space", state.Position())
						}
					} else {
						if !ascii.IsLetter(c) {
							return pars.NewError("expected sequence character", state.Position())
						}
					}

					state.Advance()
					c, err = pars.Next(state)
					j++
				}
				state.Advance()
			}

			buffer, _ := pars.Trail(state)
			gb.Origin = buffer[:len(buffer)-1]

		default:
			what := fmt.Sprintf("unexpected field name `%s`", name)
			return pars.NewError(what, state.Position())
		}
	}
}
