package gts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	ascii "gopkg.in/ascii.v1"
	pars "gopkg.in/pars.v2"
	msgpack "gopkg.in/vmihailenco/msgpack.v4"
	wrap "gopkg.in/wrap.v1"
	yaml "gopkg.in/yaml.v3"
)

// GenBankFields represents the fields of a GenBank record other than the
// features and sequence.
type GenBankFields struct {
	LocusName string `json:"locus_name" yaml:"locus_name" msgpack:"locus_name"`
	Molecule  string `json:"molecule" yaml:"molecule" msgpack:"molecule"`
	Topology  string `json:"topology" yaml:"topology" msgpack:"topology"`
	Division  string `json:"division" yaml:"division" msgpack:"division"`
	Date      Date   `json:"date" yaml:"date" msgpack:"date"`

	Definition string      `json:"definition" yaml:"definition" msgpack:"definition"`
	Accession  string      `json:"accession" yaml:"accession" msgpack:"accession"`
	Version    string      `json:"version" yaml:"version" msgpack:"version"`
	DBLink     PairList    `json:"dblink" yaml:"dblink" msgpack:"dblink"`
	Keywords   []string    `json:"keywords,omitempty" yaml:"keywords,omitempty" msgpack:"keywords,omitempty"`
	Source     Organism    `json:"source" yaml:"source" msgpack:"source"`
	References []Reference `json:"references,omitempty" yaml:"references,omitempty" msgpack:"references,omitempty"`
	Comment    string      `json:"comment" yaml:"comment" msgpack:"comment"`
}

// GenBankIO represents a temporary object for reading and writing a GenBank
// struct using various serialization libraries.
type GenBankIO struct {
	Fields   GenBankFields `json:"fields" yaml:"fields" msgpack:"fields"`
	Features []FeatureIO   `json:"features" yaml:"features" msgpack:"features"`
	Origin   []byte        `json:"origin" yaml:"origin" msgpack:"origin"`
}

// NewGenBankIO creates a new GenBankIO object.
func NewGenBankIO(gb GenBank) GenBankIO {
	fios := make([]FeatureIO, len(gb.Features))
	for i, f := range gb.Features {
		fios[i] = NewFeatureIO(f)
	}
	return GenBankIO{gb.Fields, fios, gb.Origin.Bytes()}
}

// SetGenBank sets the fields of the given GenBank pointer.
func (gbio GenBankIO) SetGenBank(gb *GenBank) error {
	gb.Fields = gbio.Fields
	gb.Origin = NewSequenceServer(New(gb.Fields, gbio.Origin))
	gb.Features = make([]Feature, len(gbio.Features))
	for i, fio := range gbio.Features {
		if err := fio.SetFeature(&gb.Features[i]); err != nil {
			return err
		}
		gb.Features[i].proxy = gb.Origin.Proxy()
	}
	return nil
}

// GenBank represents a GenBank sequence record.
type GenBank struct {
	Fields   GenBankFields
	Features FeatureList
	Origin   SequenceServer
}

func emptyGenBank() interface{} {
	return &GenBank{}
}

// EncodeWith satisfies the Encodable interface.
func (gb GenBank) EncodeWith(enc Encoder) error {
	return enc.Encode(NewGenBankIO(gb))
}

// DecodeWith satisifes the Decodable interface.
func (gb *GenBank) DecodeWith(dec Decoder) error {
	var gbio GenBankIO
	if err := dec.Decode(&gbio); err != nil {
		return err
	}
	return gbio.SetGenBank(gb)
}

// MarshalJSON satisifes the json.Marshaler interface.
func (gb GenBank) MarshalJSON() ([]byte, error) {
	return EncodeJSON(gb)
}

// UnmarshalJSON satisifes the json.Unmarshaler interface.
func (gb *GenBank) UnmarshalJSON(data []byte) error {
	return DecodeJSON(data, gb)
}

// GobEncode satisifes the gob.GobEncoder interface.
func (gb GenBank) GobEncode() ([]byte, error) {
	return EncodeGob(gb)
}

// GobDecode satisifes the gob.GobDecoder interface.
func (gb *GenBank) GobDecode(data []byte) error {
	return DecodeGob(data, gb)
}

// MarshalYAML satisifes the yaml.Marshaler interface.
func (gb GenBank) MarshalYAML() (interface{}, error) {
	return NewGenBankIO(gb), nil
}

// UnmarshalYAML satisifes the yaml.Unmarshaler interface.
func (gb *GenBank) UnmarshalYAML(value *yaml.Node) error {
	return gb.DecodeWith(value)
}

// EncodeMsgpack satisifes the msgpack.CustomEncoder interface.
func (gb GenBank) EncodeMsgpack(enc *msgpack.Encoder) error {
	return gb.EncodeWith(enc)
}

// DecodeMsgpack satisifes the msgpack.CustomDecoder interface.
func (gb *GenBank) DecodeMsgpack(dec *msgpack.Decoder) error {
	return gb.DecodeWith(dec)
}

// Info returns the metadata of the sequence.
func (gb GenBank) Info() interface{} { return gb.Fields }

// Bytes returns the byte representation of the sequence.
func (gb GenBank) Bytes() []byte { return gb.Origin.Bytes() }

// Filter the features in the list matching the selector criteria.
func (gb GenBank) Filter(ss ...FeatureFilter) []Feature {
	return gb.Features.Filter(ss...)
}

// Add a feature to the GenBank record.
func (gb GenBank) Add(f Feature) {
	f.proxy = gb.Origin.Proxy()
	gb.Features.Add(f)
}

// Insert a sequence at the specified position.
func (gb GenBank) Insert(pos int, seq Sequence) error {
	if err := gb.Origin.Insert(pos, seq); err != nil {
		return err
	}
	for _, f := range gb.Features {
		f.Location.Shift(pos, Len(seq))
	}
	return nil
}

// Delete given number of bases from the specified position.
func (gb GenBank) Delete(pos, cnt int) error {
	if err := gb.Origin.Delete(pos, cnt); err != nil {
		return err
	}
	for _, f := range gb.Features {
		f.Location.Shift(pos, -cnt)
	}
	return nil
}

// Replace the bases from the specified position with the given sequence.
func (gb GenBank) Replace(pos int, seq Sequence) error {
	return gb.Origin.Replace(pos, seq)
}

// String satisifes the fmt.Stringer interface.
func (gb GenBank) String() string {
	builder := strings.Builder{}
	indent := "            "

	length := strconv.Itoa(Len(gb))
	pad1 := strings.Repeat(" ", 28-(len(gb.Fields.LocusName)+len(length)))
	pad2 := strings.Repeat(" ", 8-len(gb.Fields.Molecule))
	pad3 := strings.Repeat(" ", 9-len(gb.Fields.Topology))
	date := strings.ToUpper(gb.Fields.Date.ToTime().Format("02-Jan-2006"))
	locus := "LOCUS       " + gb.Fields.LocusName + pad1 + length + " bp    " +
		gb.Fields.Molecule + pad2 + gb.Fields.Topology + pad3 + gb.Fields.Division +
		" " + date

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
			s := "bases"
			if gb.Fields.Molecule == "aa" {
				s = "residues"
			}
			for i, rng := range ref.Ranges {
				ranges[i] = fmt.Sprintf("%d to %d", rng.Start, rng.End)
			}
			builder.WriteString(fmt.Sprintf("(%s %s)", s, strings.Join(ranges, "; ")))
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

	ff := FeatureList(gb.Filter(Any))
	ff.Format("     ", 21).WriteTo(&builder)

	builder.WriteString("\nORIGIN      ")

	p := gb.Bytes()
	for i := 0; i < len(p); i += 60 {
		builder.WriteString(fmt.Sprintf("\n%9d ", i+1))
		for j := 0; j < 60 && i+j < len(p); j += 10 {
			k := i + j + 10
			if k > len(p) {
				k = len(p)
			}
			if j != 0 {
				builder.WriteByte(' ')
			}
			builder.Write(p[i+j : k])
		}
	}

	builder.WriteString("\n//\n")

	return builder.String()
}

// WriteTo satisfies the io.WriterTo interface.
func (gb GenBank) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, gb.String())
	return int64(n), err
}

// GenBankWriter attempts to format a record in GenBank flatfile format.
type GenBankWriter struct {
	seq Sequence
}

// WriteTo satisfies the io.WriterTo interface.
func (gf GenBankWriter) WriteTo(w io.Writer) (int64, error) {
	switch info := gf.seq.Info().(type) {
	case GenBankFields:
		return gf.WriteTo(w)
	default:
		panic(fmt.Errorf("gts does not know how to format `%T` in GenBank flatfile form", info))
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

	gb := &GenBank{Fields: fields}

	fieldNameParser := pars.Word(ascii.IsUpper).Error(errors.New("expected field name"))
	fieldBodyParser := genbankFieldBodyParser(depth)
	end := pars.Seq("//", pars.EOL).Error(errors.New("expected end of record"))

	for {
		if end(state, result) == nil {
			for i := range gb.Features {
				gb.Features[i].proxy = gb.Origin.Proxy()
			}
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

			switch v := result.Children[1].Value.(type) {
			case string:
				if v != "sites" {
					what := fmt.Sprintf("unexpected %q in reference position", v)
					return pars.NewError(what, state.Position())
				}
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
			parser := FeatureListParser("")
			if err := parser(state, result); err != nil {
				return err
			}
			ff := result.Value.([]Feature)
			gb.Features = FeatureList(ff)

		case "ORIGIN":
			pars.Line(state, result)

			origin := make([]byte, length)
			i := 0
			for i < length {
				n := []byte(strconv.Itoa(i + 1))
				m := 9 - len(n)
				for j := 0; j < m; j++ {
					c, err := pars.Next(state)
					if err != nil {
						return err
					}
					if c != ' ' {
						return pars.NewError("wanted indent", state.Position())
					}
					state.Advance()
				}
				if err := state.Request(len(n)); err != nil {
					return err
				}
				if !bytes.Equal(state.Buffer(), n) {
					return pars.NewError(fmt.Sprintf("wanted `%d`", i+1), state.Position())
				}
				state.Advance()
				c, err := pars.Next(state)
				if err != nil {
					return err
				}
				if c != ' ' {
					return pars.NewError("wanted space", state.Position())
				}
				state.Advance()
				if err := pars.Line(state, result); err != nil {
					return err
				}
				p := result.Token
				for j := 0; j < len(p); j += 11 {
					copy(origin[i:], p[j:])
					i += 10
				}
			}

			gb.Origin = NewSequenceServer(Seq(origin))

		default:
			what := fmt.Sprintf("unexpected field name `%s`", name)
			return pars.NewError(what, state.Position())
		}
	}
}

var (
	genbankJSONParser    = DecoderParser(NewJSONDecoder, emptyGenBank)
	genbankYAMLParser    = DecoderParser(NewYAMLDecoder, emptyGenBank)
	genbankMsgpackParser = DecoderParser(NewMsgpackDecoder, emptyGenBank)
)
