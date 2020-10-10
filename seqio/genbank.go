package seqio

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-ascii/ascii"
	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
	"github.com/go-wrap/wrap"
)

const defaultGenBankIndent = "            "

// FieldFormatter represents a function for formatting a field.
type FieldFormatter func(name, value string) string

// ExtraField represents an uncommon field of a genome flat-file.
type ExtraField struct {
	Name   string
	Value  string
	Format func(name, value string) string
}

// String satisfies the fmt.Stringer interface.
func (field ExtraField) String() string {
	return field.Format(field.Name, field.Value)
}

func genbankFieldFormatter(name, value string) string {
	value = AddPrefix(value, defaultGenBankIndent)
	return fmt.Sprintf("%-12s%s", name, value)
}

// GenBankExtraField creates a new extra field with a default formatter.
func GenBankExtraField(name, value string) ExtraField {
	return ExtraField{name, value, genbankFieldFormatter}
}

// GenBankFields represents the fields of a GenBank record other than the
// features and sequence.
type GenBankFields struct {
	LocusName string
	Molecule  gts.Molecule
	Topology  gts.Topology
	Division  string
	Date      Date

	Definition string
	Accession  string
	Version    string
	DBLink     Dictionary
	Keywords   []string
	Source     Organism
	References []Reference
	Comments   []string
	Extra      []ExtraField
	Contig     Contig

	Region gts.Region // Appears in sliced files.
}

// Slice returns a metadata sliced with the given region.
func (gbf GenBankFields) Slice(start, end int) interface{} {
	gbf.Region = gts.Segment{start, end}

	prefix := gbf.Molecule.Counter()
	parser := parseReferenceInfo(prefix)
	tryParse := func(info string) ([]gts.Segment, bool) {
		result, err := parser.Parse(pars.FromString(info))
		if err != nil {
			return nil, false
		}
		return result.Value.([]gts.Segment), true
	}

	refs := []Reference{}
	for _, ref := range gbf.References {
		info := ref.Info

		segs, ok := tryParse(info)
		switch {
		case ok:
			olap := []gts.Segment{}
			for _, seg := range segs {
				if seg.Overlap(start, end) {
					olap = append(olap, seg)
				}
			}
			if len(olap) > 0 {
				ss := make([]string, len(olap))
				for i, seg := range olap {
					head, tail := gts.Unpack(seg)
					head = gts.Max(0, head-start)
					tail = gts.Min(end-start, tail-start)
					ss[i] = fmt.Sprintf("%d to %d", head+1, tail)
				}
				ref.Info = fmt.Sprintf("(%s %s)", prefix, strings.Join(ss, "; "))
				refs = append(refs, ref)
			}
		default:
			refs = append(refs, ref)
		}
	}

	for i := range refs {
		refs[i].Number = i + 1
	}

	gbf.References = refs

	return gbf
}

// ID returns the ID of the sequence.
func (gbf GenBankFields) ID() string {
	if gbf.Version != "" {
		return gbf.Version
	}
	if gbf.Accession != "" {
		return gbf.Accession
	}
	return gbf.LocusName
}

// String satisifes the fmt.Stringer interface.
func (gbf GenBankFields) String() string {
	return fmt.Sprintf("%s %s", gbf.Version, gbf.Definition)
}

// GenBank represents a GenBank sequence record.
type GenBank struct {
	Fields GenBankFields
	Table  gts.FeatureTable
	Origin Origin
}

// NewGenBank creates a new GenBank object.
func NewGenBank(info GenBankFields, ff []gts.Feature, p []byte) GenBank {
	return GenBank{info, ff, NewOrigin(p)}
}

// Info returns the metadata of the sequence.
func (gb GenBank) Info() interface{} {
	return gb.Fields
}

// Features returns the feature table of the sequence.
func (gb GenBank) Features() gts.FeatureTable {
	return gb.Table
}

// Len returns the length of the sequence.
func (gb GenBank) Len() int {
	return gb.Origin.Len()
}

// Bytes returns the byte representation of the sequence.
func (gb GenBank) Bytes() []byte {
	return gb.Origin.Bytes()
}

// WithInfo creates a shallow copy of the given Sequence object and swaps the
// metadata with the given value.
func (gb GenBank) WithInfo(info interface{}) gts.Sequence {
	switch v := info.(type) {
	case GenBankFields:
		return GenBank{v, gb.Table, gb.Origin}
	default:
		return gts.New(v, gb.Features(), gb.Bytes())
	}
}

// WithFeatures creates a shallow copy of the given Sequence object and swaps
// the feature table with the given features.
func (gb GenBank) WithFeatures(ff gts.FeatureTable) gts.Sequence {
	return GenBank{gb.Fields, ff, gb.Origin}
}

// WithBytes creates a shallow copy of the given Sequence object and swaps the
// byte representation with the given byte slice.
func (gb GenBank) WithBytes(p []byte) gts.Sequence {
	return GenBank{gb.Fields, gb.Table, NewOrigin(p)}
}

// WithTopology creates a shallow copy of the given Sequence object and swaps
// the topology value with the given value.
func (gb GenBank) WithTopology(t gts.Topology) gts.Sequence {
	info := gb.Fields
	info.Topology = t
	return gb.WithInfo(info)
}

// String satisifes the fmt.Stringer interface.
func (gb GenBank) String() string {
	b := strings.Builder{}
	indent := defaultGenBankIndent

	length := gb.Origin.Len()
	if length == 0 {
		length = gb.Fields.Contig.Region.Len()
	}

	date := strings.ToUpper(gb.Fields.Date.ToTime().Format("02-Jan-2006"))
	locus := fmt.Sprintf(
		"%-12s%-17s %10d bp %6s     %-9s%s %s", "LOCUS", gb.Fields.LocusName,
		length, gb.Fields.Molecule, gb.Fields.Topology, gb.Fields.Division, date,
	)

	b.WriteString(locus + "\n")

	definition := AddPrefix(gb.Fields.Definition, indent)
	b.WriteString("DEFINITION  " + definition + ".\n")
	b.WriteString("ACCESSION   " + gb.Fields.Accession)
	if seg, ok := gb.Fields.Region.(gts.Segment); ok {
		head, tail := gts.Unpack(seg)
		b.WriteString(fmt.Sprintf(" REGION: %s", gts.Range(head, tail)))
	}
	b.WriteByte('\n')
	b.WriteString("VERSION     " + gb.Fields.Version + "\n")

	for i, pair := range gb.Fields.DBLink {
		switch i {
		case 0:
			b.WriteString("DBLINK      ")
		default:
			b.WriteString(indent)
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", pair.Key, pair.Value))
	}

	keywords := wrap.Space(strings.Join(gb.Fields.Keywords, "; ")+".", 67)
	keywords = AddPrefix(keywords, indent)
	b.WriteString("KEYWORDS    " + keywords + "\n")

	source := wrap.Space(gb.Fields.Source.Species, 67)
	source = AddPrefix(source, indent)
	b.WriteString("SOURCE      " + source + "\n")

	organism := wrap.Space(gb.Fields.Source.Name, 67)
	organism = AddPrefix(organism, indent)
	b.WriteString("  ORGANISM  " + organism + "\n")

	taxon := wrap.Space(strings.Join(gb.Fields.Source.Taxon, "; ")+".", 67)
	taxon = AddPrefix(taxon, indent)
	b.WriteString(indent + taxon + "\n")

	for _, ref := range gb.Fields.References {
		b.WriteString(fmt.Sprintf("REFERENCE   %d", ref.Number))
		if ref.Info != "" {
			pad := strings.Repeat(" ", 3-len(strconv.Itoa(ref.Number)))
			b.WriteString(pad + ref.Info)
		}
		b.WriteByte('\n')
		if ref.Authors != "" {
			b.WriteString("  AUTHORS   " + AddPrefix(ref.Authors, indent) + "\n")
		}
		if ref.Group != "" {
			b.WriteString("  CONSRTM   " + AddPrefix(ref.Group, indent) + "\n")
		}
		if ref.Title != "" {
			b.WriteString("  TITLE     " + AddPrefix(ref.Title, indent) + "\n")
		}
		if ref.Journal != "" {
			b.WriteString("  JOURNAL   " + AddPrefix(ref.Journal, indent) + "\n")
		}
		if ref.Xref != nil {
			if v, ok := ref.Xref["PUBMED"]; ok {
				b.WriteString("   PUBMED   " + v + "\n")
			}
		}
		if ref.Comment != "" {
			b.WriteString("  REMARK    " + AddPrefix(ref.Comment, indent) + "\n")
		}
	}

	for _, comment := range gb.Fields.Comments {
		b.WriteString("COMMENT     " + AddPrefix(comment, indent) + "\n")
	}

	for _, extra := range gb.Fields.Extra {
		b.WriteString(extra.String() + "\n")
	}

	b.WriteString("FEATURES             Location/Qualifiers\n")
	gb.Table.Format("     ", 21).WriteTo(&b)
	b.WriteByte('\n')

	if gb.Fields.Contig.String() != "" {
		b.WriteString(fmt.Sprintf("CONTIG      %s\n", gb.Fields.Contig))
	}

	if gb.Origin.Len() > 0 {
		b.WriteString("ORIGIN      \n")
		b.WriteString(gb.Origin.String())
	}

	b.WriteString("//\n")

	return b.String()
}

// WriteTo satisfies the io.WriterTo interface.
func (gb GenBank) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, gb.String())
	return int64(n), err
}

// GenBankFormatter implements the Formatter interface for GenBank files.
type GenBankFormatter struct {
	seq gts.Sequence
}

// WriteTo satisfies the io.WriterTo interface.
func (gf GenBankFormatter) WriteTo(w io.Writer) (int64, error) {
	switch seq := gf.seq.(type) {
	case GenBank:
		return seq.WriteTo(w)
	case *GenBank:
		return GenBankFormatter{*seq}.WriteTo(w)
	default:
		switch info := seq.Info().(type) {
		case GenBankFields:
			gb := NewGenBank(info, seq.Features(), seq.Bytes())
			return GenBankFormatter{gb}.WriteTo(w)
		default:
			return 0, fmt.Errorf("gts does not know how to format a sequence with metadata of type `%T` as GenBank", info)

		}
	}
}

var genbankLocusParser = pars.Seq(
	"LOCUS", pars.Spaces,
	pars.Word(ascii.Not(ascii.IsSpace)), pars.Spaces,
	pars.Int, pars.Any(" bp", " aa"), pars.Spaces,
	pars.Word(ascii.Not(ascii.IsSpace)), pars.Spaces,
	pars.Word(ascii.Not(ascii.IsSpace)), pars.Spaces,
	pars.Maybe(pars.Count(pars.Filter(ascii.IsUpper), 3).Map(pars.Cat)),
	pars.Spaces,
	pars.AsParser(pars.Line).Map(func(result *pars.Result) (err error) {
		s := string(result.Token)
		date, err := AsDate(s)
		result.SetValue(date)
		return err
	}),
).Children(1, 2, 4, 7, 9, 11, 13)

func tryAllParsers(pp []pars.Parser) pars.Parser {
	return func(state *pars.State, result *pars.Result) (err error) {
		for _, p := range pp {
			state.Push()
			err = p(state, result)
			if err == nil {
				state.Drop()
				return nil
			}
			if !state.Pushed() {
				return err
			}
			state.Pop()
		}
		return err
	}
}

// GenBankParser attempts to parse a single GenBank record.
func GenBankParser(state *pars.State, result *pars.Result) error {
	if err := genbankLocusParser(state, result); err != nil {
		return err
	}

	state.Clear()

	depth := len(result.Children[0].Token) + 5

	locus := string(result.Children[1].Token)
	length := result.Children[2].Value.(int)
	molecule, err := gts.AsMolecule(string(result.Children[3].Token))
	if err != nil {
		return pars.NewError(err.Error(), state.Position())
	}
	topology, err := gts.AsTopology(string(result.Children[4].Token))
	if err != nil {
		return pars.NewError(err.Error(), state.Position())
	}
	division := string(result.Children[5].Token)
	date := result.Children[6].Value.(Date)

	gb := &GenBank{Fields: GenBankFields{
		LocusName: locus,
		Molecule:  molecule,
		Topology:  topology,
		Division:  division,
		Date:      date,
		Region:    nil,
	}}

	genbankOriginParser := makeGenbankOriginParser(length)

	generators := []genbankSubparser{
		genbankDefinitionParser,
		genbankAccessionParser,
		genbankVersionParser,
		genbankDBLinkParser,
		genbankKeywordsParser,
		genbankSourceParser,
		genbankReferenceParser,
		genbankCommentParser,
		genbankFeatureParser,
		genbankContigParser,
		genbankOriginParser,
		genbankExtraFieldParser,
	}

	subparsers := make([]pars.Parser, len(generators))
	for i, generate := range generators {
		subparsers[i] = generate(gb, depth)
	}
	parser := tryAllParsers(subparsers)

	end := pars.Seq("//", pars.EOL)

	for end(state, result) != nil {
		if err := parser(state, result); err != nil {
			if dig(err) != errGenBankExtra {
				return err
			}
			pars.Line(state, result)
			if pars.End(state, result) == nil {
				return errGenBankField
			}
		}
	}

	result.SetValue(*gb)
	return nil
}
