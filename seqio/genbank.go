package seqio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-ascii/ascii"
	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
	"github.com/go-wrap/wrap"
)

const defaultGenBankIndentLength = 12

var defaultGenBankIndent = strings.Repeat(" ", defaultGenBankIndentLength)
var defaultGenBankFieldFormat = fmt.Sprintf("%%-%ds%%s\n", defaultGenBankIndentLength)

// ExtraField represents an uncommon field of a genome flat-file.
type ExtraField struct {
	Name  string
	Value string
}

// GenBankHeader represents the fields of a GenBank record other than the
// features and sequence.
type GenBankHeader struct {
	LocusName string
	Molecule  gts.Molecule
	Topology  gts.Topology
	Division  string
	Date      Date

	Definition string
	Accession  string
	Version    string
	DBLink     gts.Props
	Keywords   []string
	Source     Organism
	References []Reference
	Comments   []string
	Extra      []ExtraField

	Region gts.Region // Appears in sliced files.
}

func (header GenBankHeader) Slice(start, end int) interface{} {
	header.Region = gts.Segment{start, end}
	header.Topology = gts.Linear

	prefix := header.Molecule.Counter()
	parser := parseReferenceInfo(prefix)

	tryParse := func(info string) ([]gts.Ranged, bool) {
		result, err := parser.Parse(pars.FromString(info))
		if err != nil {
			return nil, false
		}
		return result.Value.([]gts.Ranged), true
	}

	refs := []Reference{}
	for _, ref := range header.References {
		info := ref.Info

		locs, ok := tryParse(info)
		switch {
		case ok:
			olap := []gts.Ranged{}
			for _, loc := range locs {
				if gts.LocationOverlap(loc, start, end) {
					olap = append(olap, loc)
				}
			}
			if len(olap) > 0 {
				ss := make([]string, len(olap))
				for i, loc := range olap {
					head, tail := loc.Start, loc.End
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

	header.References = refs

	return header
}

func toOriginLength(length int) int {
	lines := length / 60
	ret := lines * 76

	lastLine := length % 60

	if lastLine == 0 {
		return ret
	}

	blocks := lastLine / 10
	ret += 10 + blocks*11

	lastBlock := lastLine % 10
	if lastBlock == 0 {
		return ret
	}

	return ret + lastBlock + 1
}

func fromOriginLength(length int) int {
	lines := length / 76
	ret := lines * 60

	lastLine := length % 76
	if lastLine == 0 {
		return ret
	}

	lastLine -= 11
	blocks := lastLine / 11
	return ret + (blocks * 10) + (lastLine % 11)
}

// Origin represents a GenBank sequence origin value.
type Origin struct {
	Buffer []byte
	Parsed bool
}

// NewOrigin formats a byte slice into GenBank sequence origin format.
func NewOrigin(p []byte) *Origin {
	length := len(p)
	q := make([]byte, toOriginLength(length))
	offset := 0
	for i := 0; i < length; i += 60 {
		prefix := fmt.Sprintf("%9d", i+1)
		offset += copy(q[offset:], prefix)
		for j := 0; j < 60 && i+j < length; j += 10 {
			start := i + j
			end := gts.Min(i+j+10, length)
			q[offset] = spaceByte
			offset++
			offset += copy(q[offset:], p[start:end])
		}
		q[offset] = '\n'
		offset++
	}
	return &Origin{q, false}
}

// Bytes converts the GenBank sequence origin into a byte slice.
func (o *Origin) Bytes() []byte {
	if !o.Parsed {
		p := o.Buffer
		if len(p) < 12 {
			return nil
		}

		length := fromOriginLength(len(p))
		q := make([]byte, length)
		offset, start := 0, 0
		for i := 0; i < length; i += 60 {
			start += 9
			for j := 0; j < 60 && i+j < length; j += 10 {
				start++
				end := gts.Min(start+10, len(p)-1)
				offset += copy(q[offset:], p[start:end])
				start = end
			}
			start++
		}

		o.Buffer = q
		o.Parsed = true
	}

	return o.Buffer
}

// String satisfies the fmt.Stringer interface.
func (o Origin) String() string {
	if !o.Parsed {
		return string(o.Buffer)
	}
	return string(NewOrigin(o.Buffer).Buffer)
}

// Len returns the actual sequence length.
func (o Origin) Len() int {
	if len(o.Buffer) == 0 {
		return 0
	}
	if o.Parsed {
		return len(o.Buffer)
	}
	return fromOriginLength(len(o.Buffer))
}

// Contig represents a contig field.
type Contig struct {
	Accession string
	Region    gts.Segment
}

// Len returns the length of the sequence.
func (contig Contig) Len() int {
	return contig.Region.Len()
}

// String satisfies the fmt.Stringer interface.
func (contig Contig) String() string {
	if contig.Accession == "" {
		return ""
	}
	head, tail := gts.Unpack(contig.Region)
	return fmt.Sprintf("join(%s:%d..%d)", contig.Accession, head+1, tail)
}

// Bytes returns the byte representation of the sequence.
func (contig Contig) Bytes() []byte {
	gts.Errorln("GenBank Contigs do not contain sequence reprensentations")
	panic("invalid operation")
}

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
			what := fmt.Sprintf("uneven indent for field `%s`", name)
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

func genbankExtraFieldParser(header *GenBankHeader, depth int) pars.Parser {
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
		extra := ExtraField{name, value}
		header.Extra = append(header.Extra, extra)
		return nil
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

func genbankDefinitionParser(header *GenBankHeader, depth int) pars.Parser {
	parser := genbankGenericFieldParser("DEFINITION", depth)
	return parser.Map(func(result *pars.Result) error {
		p := result.Token
		if len(p) != 0 && p[len(p)-1] != '.' {
			return errors.New("expected period")
		}
		p = bytes.TrimSuffix(p, []byte{'.'})
		header.Definition = string(p)
		return nil
	})
}

func tryParseGenBankAccessionRegion(s string) (string, gts.Region) {
	index := strings.Index(s, " REGION: ")
	if index < 0 {
		return s, nil
	}
	loc, err := gts.AsLocation(s[index+9:])
	if err != nil {
		return s, nil
	}
	r, ok := loc.Region().(gts.Segment)
	if !ok {
		return s, nil
	}
	return s[:index], r
}

func genbankAccessionParser(header *GenBankHeader, depth int) pars.Parser {
	parser := genbankGenericFieldParser("ACCESSION", depth)
	return parser.Map(func(result *pars.Result) error {
		acc, r := tryParseGenBankAccessionRegion(string(result.Token))
		header.Accession = acc
		header.Region = r
		return nil
	})
}

func genbankVersionParser(header *GenBankHeader, depth int) pars.Parser {
	fieldParser := genbankGenericFieldParser("VERSION", depth)
	return fieldParser.Map(func(result *pars.Result) error {
		header.Version = string(result.Token)
		return nil
	})
}

func genbankDBLinkPairParser(header *GenBankHeader, depth int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		pars.Line(state, result)
		s := string(result.Token)
		switch i := strings.IndexByte(s, ':'); i {
		case -1:
			return pars.NewError("expected `:`", state.Position())
		default:
			db, id := s[:i], s[i+2:]
			header.DBLink.Set(db, id)
			return nil
		}
	}
}

func genbankDBLinkParser(header *GenBankHeader, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("DBLINK", depth)
	indentParser := pars.String(strings.Repeat(" ", depth))
	pairParser := genbankDBLinkPairParser(header, depth)
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

func genbankKeywordsParser(header *GenBankHeader, depth int) pars.Parser {
	fieldNameParser := genbankFieldNameParser("KEYWORDS", depth)
	fieldBodyParser := genbankFieldBodyParser(depth, ' ')
	return func(state *pars.State, result *pars.Result) error {
		if err := fieldNameParser(state, pars.Void); err != nil {
			return err
		}
		fieldBodyParser(state, result)
		header.Keywords = INSDCSplit(string(result.Token))
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

func genbankSourceParser(header *GenBankHeader, depth int) pars.Parser {
	sourceParser := genbankGenericFieldParser("SOURCE", depth)
	sourceParser = sourceParser.Map(func(result *pars.Result) error {
		header.Source.Species = string(result.Token)
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
		header.Source.Name = string(result.Token)

		w := bytes.Buffer{}
		for fieldLineParser(state, result) == nil {
			if len(result.Token) > 0 {
				if w.Len() > 0 {
					w.WriteByte(spaceByte)
				}
				w.Write(result.Token)
			}
		}
		header.Source.Taxon = INSDCSplit(w.String())

		return nil
	}
}

func genbankReferenceSubfieldParser(ref *Reference, depth int) pars.Parser {
	authorsParser := genbankGenericSubfieldParser("AUTHORS", depth)
	consrtmParser := genbankGenericSubfieldParser("CONSRTM", depth)
	titleParser := genbankGenericSubfieldParser("TITLE", depth)
	journalParser := genbankGenericSubfieldParser("JOURNAL", depth)
	pubmedParser := genbankGenericSubfieldParser("PUBMED", depth)
	remarkParser := genbankGenericSubfieldParser("REMARK", depth)
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

func genbankReferenceParser(header *GenBankHeader, depth int) pars.Parser {
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

		header.References = append(header.References, ref)

		return nil
	}
}

func genbankCommentParser(header *GenBankHeader, depth int) pars.Parser {
	fieldParser := genbankGenericFieldParser("COMMENT", depth)
	return fieldParser.Map(func(result *pars.Result) error {
		header.Comments = append(header.Comments, string(result.Token))
		return nil
	})
}

func genbankFeatureParser(ff *gts.Features) pars.Parser {
	featureNameParser := pars.String("FEATURES")
	featureBodyParser := INSDCTableParser("")
	return func(state *pars.State, result *pars.Result) error {
		if err := featureNameParser(state, result); err != nil {
			return err
		}
		pars.Line(state, result)
		state.Clear()
		if err := featureBodyParser(state, result); err != nil {
			return err
		}
		*ff = result.Value.(gts.Features)
		return nil
	}
}

func genbankContigParser(seq *gts.Sequence, depth int) pars.Parser {
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
		*seq = Contig{accession, gts.Segment{head - 1, tail}}
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

func makeGenbankOriginParser(seq *gts.Sequence, length, depth int) pars.Parser {
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
			*seq = &Origin{p, false}
			return nil
		}

		parser := slowGenBankOriginParser(length)
		if err := parser(state, result); err != nil {
			return err
		}
		p = result.Token
		*seq = &Origin{p, false}
		return nil
	}
}

func GenBankParser(state *pars.State, result *pars.Result) error {
	state.Push()
	if err := genbankLocusParser(state, result); err != nil {
		state.Pop()
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

	header := GenBankHeader{
		LocusName: locus,
		Molecule:  molecule,
		Topology:  topology,
		Division:  division,
		Date:      date,
		Region:    nil,
	}

	ff := gts.Features{}

	var seq gts.Sequence

	parser := pars.Any(
		genbankDefinitionParser(&header, depth),
		genbankAccessionParser(&header, depth),
		genbankVersionParser(&header, depth),
		genbankDBLinkParser(&header, depth),
		genbankKeywordsParser(&header, depth),
		genbankSourceParser(&header, depth),
		genbankReferenceParser(&header, depth),
		genbankCommentParser(&header, depth),
		genbankFeatureParser(&ff),
		genbankContigParser(&seq, depth),
		makeGenbankOriginParser(&seq, length, depth),
		genbankExtraFieldParser(&header, depth),
	)
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

	result.SetValue(Record{
		Header:   header,
		Features: ff,
		Sequence: seq,
	})

	return nil
}

type GenBankIStream struct {
	state  *pars.State
	result *pars.Result
	index  int
	peeked bool
}

func NewGenBankIStream(r io.Reader) *GenBankIStream {
	return &GenBankIStream{
		state:  pars.NewState(r),
		result: &pars.Result{},
		index:  0,
		peeked: false,
	}
}

func (istream *GenBankIStream) Peek() error {
	if istream.peeked {
		return nil
	}
	istream.state.Push()
	if err := GenBankParser(istream.state, istream.result); err != nil {
		istream.state.Drop()
		return err
	}
	istream.peeked = true
	return nil
}

func (istream *GenBankIStream) Next(fh FeatureHandler) error {
	if err := istream.Peek(); err != nil {
		return err
	}
	rec := istream.result.Value.(Record)
	if err := rec.Manipulate(fh, istream.index); err != nil {
		return err
	}
	istream.index++
	istream.peeked = false
	return nil
}

func (istream *GenBankIStream) ForEach(fh FeatureHandler) error {
	var err error
	for err == nil {
		err = istream.Next(fh)
	}
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return err
}

type GenBankOStream struct {
	w  io.Writer
	hh []GenBankHeader
	ff []gts.Features
	ss []gts.Sequence
}

func NewGenBankOstream(w io.Writer) OStream {
	return &GenBankOStream{
		w:  w,
		hh: nil,
		ff: nil,
		ss: nil,
	}
}

func (ostream *GenBankOStream) PushHeader(header interface{}) error {
	switch v := header.(type) {
	case GenBankHeader:
		ostream.hh = append(ostream.hh, v)
		return ostream.tryWrite()
	default:
		return fmt.Errorf("gts does not know how to format a sequence with header type `%T` as GenBank", v)
	}
}

func (ostream *GenBankOStream) PushFeatures(ff gts.Features) error {
	ostream.ff = append(ostream.ff, ff)
	return ostream.tryWrite()
}

func (ostream *GenBankOStream) PushSequence(seq gts.Sequence) error {
	if contig, ok := seq.(Contig); ok && contig.Accession == "" {
		return errors.New("invalid CONTIG")

	}
	ostream.ss = append(ostream.ss, seq)
	return ostream.tryWrite()
}

func (ostream *GenBankOStream) tryWrite() error {
	if len(ostream.hh) == 0 || len(ostream.ff) == 0 || len(ostream.ss) == 0 {
		return nil
	}

	header, ff, seq := ostream.hh[0], ostream.ff[0], ostream.ss[0]

	ostream.hh = ostream.hh[1:]
	ostream.ff = ostream.ff[1:]
	ostream.ss = ostream.ss[1:]

	sources := ff.Filter(gts.Key("source"))
	others := ff.Filter(gts.Not(gts.Key("source")))
	for i, f := range sources {
		sources[i].Loc = gts.AsComplete(f.Loc)
	}
	ff = append(sources, others...)

	b := bytes.Buffer{}

	b.WriteString(fmt.Sprintf(
		"%-12s%-17s %10d bp %6s     %-9s%s %s\n", "LOCUS",
		header.LocusName,
		gts.Len(seq),
		header.Molecule,
		header.Topology,
		header.Division,
		strings.ToUpper(header.Date.ToTime().Format("02-Jan-2006")),
	))

	b.WriteString(fmtGenBankField("DEFINITION", header.Definition+"."))

	acc := header.Accession
	if seg, ok := header.Region.(gts.Segment); ok {
		acc = fmt.Sprintf("%s REGION: %s", acc, gts.Range(gts.Unpack(seg)))
	}
	b.WriteString(fmtGenBankField("ACCESSION", acc))

	b.WriteString(fmt.Sprintf("VERSION     %s\n", header.Version))

	ss := make([]string, header.DBLink.Len())
	for i, pair := range header.DBLink.Items() {
		ss[i] = fmt.Sprintf("%s: %s", pair.Key, pair.Value)
	}
	dblink := strings.Join(ss, "\n")
	b.WriteString(fmtGenBankField("DBLINK", dblink))

	b.WriteString(fmtGenBankField("KEYWORDS", INSDCJoin(header.Keywords)))

	if header.Source.Species != "" {
		b.WriteString(fmtGenBankField("SOURCE", header.Source.Species))
	}

	if header.Source.Name != "" {
		b.WriteString(fmtGenBankField("  ORGANISM", header.Source.Name))
	}

	if len(header.Source.Taxon) > 0 {
		b.WriteString(fmtGenBankField("", INSDCJoin(header.Source.Taxon)))
	}

	for _, ref := range header.References {
		number := strconv.Itoa(ref.Number)
		if ref.Info != "" {
			number = fmt.Sprintf("%-3s%s", number, ref.Info)
		}
		b.WriteString(fmtGenBankField("REFERENCE", number))

		if ref.Authors != "" {
			b.WriteString(fmtGenBankField("  AUTHORS", ref.Authors))
		}

		if ref.Group != "" {
			b.WriteString(fmtGenBankField("  CONSRTM", ref.Group))
		}

		if ref.Title != "" {
			b.WriteString(fmtGenBankField("  TITLE", ref.Title))
		}

		if ref.Journal != "" {
			b.WriteString(fmtGenBankField("  JOURNAL", ref.Journal))
		}

		if ref.Xref != nil {
			if xref, ok := ref.Xref["PUBMED"]; ok {
				b.WriteString(fmtGenBankField("   PUBMED", xref))
			}
		}

		if ref.Comment != "" {
			b.WriteString(fmtGenBankField("  REMARK", ref.Comment))
		}
	}

	for _, comment := range header.Comments {
		b.WriteString(
			fmt.Sprintf(
				defaultGenBankFieldFormat,
				"COMMENT",
				AddPrefix(comment, defaultGenBankIndent),
			),
		)
	}

	for _, extra := range header.Extra {
		b.WriteString(fmtGenBankField(extra.Name, extra.Value))
	}

	b.WriteString("FEATURES             Location/Qualifiers\n")
	INSDCFormatter{ff, "     ", 21}.WriteTo(&b)
	b.WriteByte('\n')

	switch seq := seq.(type) {
	case Contig:
		head, tail := gts.Unpack(seq.Region)
		contig := fmt.Sprintf("join(%s:%d..%d)", seq.Accession, head+1, tail)
		b.WriteString(fmtGenBankField("CONTIG", contig))

	case *Origin:
		b.WriteString(fmtGenBankField("ORIGIN", ""))
		b.WriteString(seq.String())

	default:
		b.WriteString(fmtGenBankField("ORIGIN", ""))
		b.WriteString(NewOrigin(seq.Bytes()).String())
	}

	b.WriteString("//\n")

	_, err := ostream.w.Write(b.Bytes())
	return err
}

func fmtGenBankField(name string, value string) string {
	value = wrap.Space(value, 79-defaultGenBankIndentLength)
	value = AddPrefix(value, defaultGenBankIndent)
	return fmt.Sprintf(defaultGenBankFieldFormat, name, value)
}
