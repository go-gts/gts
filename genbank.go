package gts

import (
	"errors"
	"fmt"
	"strings"
	"time"

	ascii "gopkg.in/ktnyt/ascii.v1"
	pars "gopkg.in/ktnyt/pars.v2.4"
)

// GenBankFields represents the fields of a GenBank record other than the
// features and sequence.
type GenBankFields struct {
	LocusName string
	Molecule  string
	Topology  string
	Division  string
	Date      time.Time

	Definition string
	Accession  string
	Version    string
	DBLink     PairList
	Keywords   []string
	Source     Organism
	References []Reference
	Comment    string
}

// GenBank represents a GenBank sequence record.
type GenBank struct {
	Fields   GenBankFields
	features FeatureTable
	origin   []byte
}

// Bytes satisfies the gts.Sequence interface.
func (gb GenBank) Bytes() []byte { return gb.origin }

// Metadata returns the metadata of the GenBank record.
func (gb GenBank) Metadata() interface{} { return gb.Fields }

// Features returns the feature table of the GenBank record.
func (gb GenBank) Features() FeatureTable { return gb.features }

var genbankLocusParser = pars.Seq(
	"LOCUS", pars.Spaces,
	pars.Word(ascii.IsSnake), pars.Spaces,
	pars.Int, " bp", pars.Spaces,
	pars.Word(ascii.Not(ascii.IsSpace)), pars.Spaces,
	pars.Any("linear", "circular"), pars.Spaces,
	pars.Count(pars.Byte(), 3), pars.Spaces,
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

// GenBankParser will attempt to parse a single GenBank record.
func GenBankParser(state *pars.State, result *pars.Result) error {
	if err := genbankLocusParser(state, result); err != nil {
		return err
	}

	depth := len(result.Children[0].Token) + 5
	indent := pars.String(strings.Repeat(" ", depth))

	locus := string(result.Children[1].Token)
	length := result.Children[2].Value.(int)
	molecule := string(result.Children[3].Token)
	topology := result.Children[4].Value.(string)
	division := string(result.Children[5].Token)
	date := result.Children[6].Value.(time.Time)

	fields := GenBankFields{
		LocusName: locus,
		Molecule:  molecule,
		Topology:  topology,
		Division:  division,
		Date:      date,
	}

	gb := &GenBank{Fields: fields}

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
			id := string(result.Children[1].Token)
			gb.Fields.DBLink.Set(db, id)

			tailParser := pars.Many(pars.Seq(indent, headParser).Child(1))
			tailParser(state, result)
			for _, child := range result.Children {
				db = string(child.Children[0].Token)
				id = string(child.Children[1].Token)
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
						"bases ",
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

			switch v := result.Children[1].Value.(type) {
			case string:
				if v != "sites" {
					what := fmt.Sprintf("unexpected %q in reference position", v)
					return pars.NewError(what, state.Position())
				}
				reference.Ranges = []Range{}
			default:
				children := result.Children[1].Children
				ranges := make([]Range, len(children))
				for i, child := range children {
					start := child.Children[0].Value.(int)
					end := child.Children[1].Value.(int)
					ranges[i] = Range{start, end}
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
			gb.features = result.Value.(FeatureTable)

		case "ORIGIN":
			pars.Line(state, result)
			seqlineParser := pars.Seq(pars.Spaces, pars.Int, ' ', pars.Line)

			origin := make([]byte, length)

			i := 0
			for i < length {
				if err := seqlineParser(state, result); err != nil {
					return err
				}
				n := result.Children[1].Value.(int)
				if i+1 != n {
					return pars.NewError("number of bases does not match", state.Position())
				}
				p := result.Children[3].Token
				for j := 0; j < len(p); j += 11 {
					copy(origin[i:], p[j:])
					i += 10
				}
			}

			gb.origin = origin

		default:
			what := fmt.Sprintf("unexpected field name `%s`", name)
			return pars.NewError(what, state.Position())
		}
	}
}
