package seqio

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ktnyt/gods"
	"github.com/ktnyt/gt1"
	"github.com/ktnyt/pars"
)

var genbankFieldDepth = 12
var genbankFeatureIndent = 5
var genbankFeatureDepth = 21

func formatLocusGenBank(record *gt1.Record) string {
	metadata := record.Metadata()
	name := metadata.LocusName
	length := strconv.Itoa(record.Len())
	pad1 := strings.Repeat(" ", 28-(len(name)+len(length)))
	molecule := metadata.Molecule
	pad2 := strings.Repeat(" ", 8-len(molecule))
	geometry := metadata.Topology
	pad3 := strings.Repeat(" ", 9-len(geometry))
	division := metadata.Division
	date := strings.ToUpper(metadata.Dates[0].Format("02-Jan-2006"))
	return "LOCUS       " + name + pad1 + length + " bp    " + molecule + pad2 + geometry + pad3 + division + " " + date
}

func formatSourceGenBank(source gt1.Organism) string {
	prefix := strings.Repeat(" ", genbankFieldDepth)
	wrapSpace := WrapAt(' ', 80, prefix)

	lines := make([]string, 0, 3)
	lines = append(lines, wrapSpace("SOURCE      "+source.Species))
	lines = append(lines, wrapSpace("  ORGANISM  "+source.Name))

	if source.Taxon != nil {
		lines = append(lines, wrapSpace("            "+strings.Join(source.Taxon, "; ")+"."))
	}

	return strings.Join(lines, "\n")
}

func formatReferenceGenBank(reference gt1.Reference) string {
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf(
		"REFERENCE   %-2d (bases %d to %d)",
		reference.Number, reference.Start, reference.End,
	))

	prefix := strings.Repeat(" ", genbankFieldDepth)
	wrapSpace := WrapAt(' ', 80, prefix)

	if reference.Authors != "" {
		lines = append(lines, wrapSpace("  AUTHORS   "+reference.Authors))
	}

	if reference.Group != "" {
		lines = append(lines, wrapSpace("  CONSRTM   "+reference.Group))
	}

	if reference.Title != "" {
		lines = append(lines, wrapSpace("  TITLE     "+reference.Title))
	}

	if reference.Journal != "" {
		lines = append(lines, wrapSpace("  JOURNAL   "+reference.Journal))
	}

	if reference.Xref != nil {
		if v, ok := reference.Xref["PUBMED"]; ok {
			lines = append(lines, wrapSpace("   PUBMED   "+v))
		}
	}

	if reference.Comment != "" {
		lines = append(lines, wrapSpace("  REMARK    "+reference.Comment))
	}

	return strings.Join(lines, "\n")
}

func formatFeatureGenBank(feature *gt1.Feature) string {
	lines := make([]string, 0)
	featureKey := strings.Repeat(" ", genbankFeatureIndent) + feature.Key() + strings.Repeat(" ", genbankFeatureDepth-(genbankFeatureIndent+len(feature.Key())))
	lines = append(lines, featureKey+feature.Location().Format())

	for _, pair := range feature.Qualifiers().Iter() {
		property := strings.Repeat(" ", genbankFeatureDepth)
		key, value := pair.Key, pair.Value
		if n, err := strconv.Atoi(value); err == nil {
			property += fmt.Sprintf("/%s=%d", key, n)
		} else if key == "rpt_type" || key == "transl_except" {
			property += fmt.Sprintf("/%s=%s", key, value)
		} else {
			if len(value) > 0 {
				property += fmt.Sprintf("/%s=\"%s\"", key, value)
			} else {
				property += fmt.Sprintf("/%s", key)
			}
		}

		prefix := strings.Repeat(" ", genbankFeatureDepth)
		wrap := Wrap(80, prefix)
		wrapSpace := WrapAt(' ', 80, prefix)

		if key == "translation" {
			property = wrap(property)
		} else {
			property = wrapSpace(property)
		}
		lines = append(lines, property)
	}
	return strings.Join(lines, "\n")
}

func FormatGenBank(record *gt1.Record) string {
	prefix := strings.Repeat(" ", genbankFieldDepth)
	wrapSpace := WrapAt(' ', 80, prefix)

	m := record.Metadata()

	lines := make([]string, 0)
	lines = append(lines, formatLocusGenBank(record))
	lines = append(lines, wrapSpace("DEFINITION  "+m.Definition))
	lines = append(lines, "ACCESSION   "+m.Accessions[0])
	lines = append(lines, "VERSION     "+m.Version)

	if m.DBLink.Len() > 0 {
		for i, pair := range m.DBLink.Iter() {
			dblink := strings.Repeat(" ", genbankFieldDepth)
			if i == 0 {
				dblink = "DBLINK" + strings.Repeat(" ", genbankFieldDepth-6)
			}
			dblink += fmt.Sprintf("%s: %s", pair.Key, pair.Value)
			lines = append(lines, dblink)
		}
	}

	lines = append(lines, wrapSpace("KEYWORDS    "+strings.Join(m.Keywords, "; ")+"."))
	lines = append(lines, formatSourceGenBank(m.Source))

	for _, reference := range m.References {
		lines = append(lines, formatReferenceGenBank(reference))
	}

	if len(m.Comment) > 0 {
		lines = append(lines, wrapSpace("COMMENT     "+m.Comment))
	}

	lines = append(lines, "FEATURES             Location/Qualifiers")
	for _, feature := range record.Features().Iter() {
		lines = append(lines, formatFeatureGenBank(feature))
	}
	lines = append(lines, "ORIGIN      ")
	for i := 0; i < record.Len(); i += 60 {
		seq := make([]string, 0, 6)
		for j := 0; j < 60 && i+j < record.Len(); j += 10 {
			k := i + j + 10
			if i+j+10 > record.Len() {
				k = record.Len()
			}
			seq = append(seq, record.Slice(i+j, k).String())
		}
		lines = append(lines, fmt.Sprintf("%9d %s", i+1, strings.Join(seq, " ")))
	}
	lines = append(lines, "//")
	return strings.Join(lines, "\n")
}

type genbankFieldName struct {
	Indent int
	Value  string
	Depth  int
}

var genbankFieldNameParser = pars.Seq(
	pars.Many(' '),
	pars.UpperWord.Map(pars.CatByte),
	pars.Many(' '),
).Map(func(result *pars.Result) error {
	indent := len(result.Children[0].Children)
	value := result.Children[1].Value.(string)
	depth := indent + len(value) + len(result.Children[2].Children)
	result.Value = genbankFieldName{Indent: indent, Value: value, Depth: depth}
	result.Children = nil
	return nil
})

func genbankFieldBodyParser(indent, depth int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		// Remove leading spaces.
		if err := pars.Many(' ')(state, result); err != nil {
			return pars.NewTraceError("GenBank Field Body", err)
		}

		// The first line should be available.
		if err := pars.Line(state, result); err != nil {
			return pars.NewTraceError("GenBank Field Body", err)
		}
		body := result.Value.(string)

		// Keep reading lines with same depth.
		for {
			// Count the number of leading spaces.
			count := 0
			state.Mark()
			if err := state.Want(1); err != nil {
				state.Jump()
				return err
			}
			for state.Buffer[state.Index] == ' ' {
				state.Advance(1)
				count += 1
				if err := state.Want(1); err != nil {
					state.Jump()
					return err
				}
			}

			// Append the line to the body text.
			// This must be processed first so the rest of the code does not mistake a
			// valid body line for a subfield.
			// Add a space first to accomodate for the indent.
			if depth == count {
				state.Unmark()
				if err := pars.Line(state, result); err != nil {
					return pars.NewTraceError("GenBank Field Body", err)
				}
				body += " " + result.Value.(string)
			} else {
				// Found shallower indent so return.
				if count <= indent {
					result.Value = body
					result.Children = nil
					state.Jump()
					return nil
				}

				// Mismatching depth is not currently tolerated unless if it can be a subfield.
				// This bit introduces backtracking which slightly hinders performance.
				// Although, with the current GenBank specifications it is impossible to optimize.
				if err := genbankFieldNameParser(state, pars.VoidResult); err != nil {
					state.Jump()
					return pars.NewTraceError("GenBank Field Body", err)
				} else {
					result.Value = body
					result.Children = nil
					state.Jump()
					return nil
				}

				state.Jump()
				return pars.NewMismatchError("GenBank Field Body", []byte("matching depth"), state.Position)
			}
		}
	}
}

var genbankLocusParser = pars.Phrase(
	pars.WordLike(notFilter(pars.IsWhitespace)),
	pars.Integer.Map(pars.Atoi), "bp",
	pars.Word,
	pars.Word,
	pars.Word,
	pars.AsParser(pars.Line).Map(pars.Time("02-Jan-2006")),
).Map(pars.Children(0, 1, 3, 4, 5, 6))

var genbankDBLinkEntryParser = pars.Seq(
	pars.WordLike(notByte(':')).Map(pars.CatByte), ": ",
	pars.WordLike(notByte('\n')).Map(pars.CatByte), '\n',
).Map(pars.Children(0, 2))

func genbankDBLinkParser(depth int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		dblink := gods.NewOrdered()
		for {
			if err := genbankDBLinkEntryParser(state, result); err != nil {
				return pars.NewTraceError("GenBank DBLink", err)
			}

			dblink.Add(result.Children[0].Value.(string), result.Children[1].Value.(string))

			if state.Buffer[state.Index] != ' ' {
				result.Value = dblink
				result.Children = nil
				return nil
			}

			state.Mark()
			for state.Buffer[state.Index] == ' ' {
				if err := state.Want(1); err != nil {
					state.Jump()
					return pars.NewTraceError("GenBank DBLink", err)
				}
				state.Advance(1)
			}
			state.Unmark()
		}
	}
}

func genbankSourceParser(depth int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		// First process the source line.
		if err := pars.Line(state, result); err != nil {
			return err
		}
		source := gt1.Organism{Species: result.Value.(string)}

		if err := genbankFieldNameParser(state, result); err != nil {
			return pars.NewTraceError("GenBank Source", err)
		}

		// Process the ORGANISM line.
		fieldName := result.Value.(genbankFieldName)
		if fieldName.Value != "ORGANISM" {
			return pars.NewMismatchError("GenBank Source", []byte("ORGANISM"), state.Position)
		}

		if fieldName.Depth != depth {
			return pars.NewMismatchError("GenBank Source", []byte("matching depth"), state.Position)
		}

		if err := pars.Line(state, result); err != nil {
			return pars.NewTraceError("GenBank Source", err)
		}
		source.Name = result.Value.(string)

		// Parse taxonomy like other GenBank fields.
		if err := genbankFieldBodyParser(0, depth)(state, result); err != nil {
			return pars.NewTraceError("GenBank Source", err)
		}
		source.Taxon = FlatFileSplit(result.Value.(string))

		result.Value = source
		result.Children = nil
		return nil
	}
}

var genbankReferenceRangeParser = pars.Phrase(
	pars.Integer.Map(pars.Atoi),
	"(bases", pars.Integer.Map(pars.Atoi), "to", pars.Integer.Map(pars.Atoi), ')',
).Map(pars.Children(0, 2, 4))

func genbankReferenceParser(depth int) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		// Parse the reference range first.
		if err := genbankReferenceRangeParser(state, result); err != nil {
			return pars.NewTraceError("GenBank Reference", err)
		}
		number := result.Children[0].Value.(int)
		start := result.Children[1].Value.(int)
		end := result.Children[2].Value.(int)
		pars.Try('\n')(state, result)

		reference := gt1.Reference{
			Number: number,
			Start:  start,
			End:    end,
		}

		// Find all subfields.
		indent := -1
		for {
			state.Mark()
			if err := genbankFieldNameParser(state, result); err != nil {
				state.Jump()
				return pars.NewTraceError("GenBank Reference", err)
			}
			name := result.Value.(genbankFieldName)

			if indent < 0 {
				indent = name.Indent
			}

			if name.Indent < indent {
				state.Jump()
				result.Value = reference
				result.Children = nil
				return nil
			}

			if err := genbankFieldBodyParser(indent, depth)(state, result); err != nil {
				state.Jump()
				return pars.NewTraceError("GenBank Reference", err)
			}
			body := result.Value.(string)
			switch name.Value {
			case "AUTHORS":
				reference.Authors = body
			case "CONSRTM":
				reference.Group = body
			case "TITLE":
				reference.Title = body
			case "JOURNAL":
				reference.Journal = body
			case "PUBMED":
				reference.Xref = map[string]string{"PUBMED": body}
			case "REMARK":
				reference.Comment = body
			}
			state.Unmark()
		}
	}
}

var genbankOriginLineParser = pars.Seq(
	pars.Many(' '),
	pars.Integer,
	' ',
	pars.Line,
).Map(pars.Child(3)).Map(func(result *pars.Result) error {
	value := result.Value.(string)
	value = strings.Replace(value, " ", "", 5)
	result.Value = value
	result.Children = nil
	return nil
})

func GenBankParser(state *pars.State, result *pars.Result) error {
	if err := genbankFieldNameParser(state, result); err != nil {
		return pars.NewTraceError("GenBank", err)
	}

	fieldName := result.Value.(genbankFieldName)

	// The first field must be a LOCUS.
	if fieldName.Value != "LOCUS" {
		return pars.NewMismatchError("GenBank", []byte("LOCUS"), state.Position)
	}

	locusDepth := fieldName.Depth
	fieldBodyParser := genbankFieldBodyParser(0, locusDepth)

	if err := genbankLocusParser(state, result); err != nil {
		return pars.NewTraceError("GenBank", err)
	}

	length := result.Children[1].Value.(int)
	locusName := result.Children[0].Value.(string)
	molecule := result.Children[2].Value.(string)
	topology := result.Children[3].Value.(string)
	division := result.Children[4].Value.(string)
	date := result.Children[5].Value.(time.Time)

	metadata := &gt1.Metadata{
		LocusName: locusName,
		Molecule:  molecule,
		Topology:  topology,
		Division:  division,
		Dates:     []time.Time{date},
	}

	features := &gt1.FeatureTable{}
	sequence := make([]byte, length)

	pars.Try('\n')(state, result)

	// Continually process fields.
	for {
		// End of entry.
		if err := pars.AsParser("//")(state, result); err == nil {
			pars.Try(pars.Many(pars.Space))(state, result)
			result.Value = gt1.NewRecord(metadata, features, sequence)
			result.Children = nil
			return nil
		}

		if err := genbankFieldNameParser(state, result); err != nil {
			return pars.NewTraceError("GenBank", err)
		}
		fieldName = result.Value.(genbankFieldName)

		// FEATURES and ORIGIN do not fit the field conventions.
		if fieldName.Value == "FEATURES" {
			if err := gt1.FeatureTableParser(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			features = result.Value.(*gt1.FeatureTable)
			continue
		}

		if fieldName.Value == "ORIGIN" {
			pars.Try('\n')(state, result)
			count := 0

			for state.Buffer[state.Index] == ' ' {
				if err := genbankOriginLineParser(state, result); err != nil {
					return pars.NewTraceError("GenBank", err)
				}
				line := []byte(result.Value.(string))
				copy(sequence[count:], line)
				count += len(line)
				state.Clear()
				state.Want(1)
			}
			continue
		}

		if fieldName.Depth != locusDepth {
			return pars.NewMismatchError("GenBank", []byte("matching field depth"), state.Position)
		}

		// Parse the specialized fields.
		switch fieldName.Value {
		case "DEFINITION":
			if err := fieldBodyParser(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.Definition = result.Value.(string)
		case "ACCESSION":
			if err := pars.Line(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.Accessions = []string{result.Value.(string)}
		case "VERSION":
			if err := pars.Line(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.Version = result.Value.(string)
		case "DBLINK":
			if err := genbankDBLinkParser(locusDepth)(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.DBLink = result.Value.(*gods.Ordered)
		case "KEYWORDS":
			if err := pars.Line(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.Keywords = FlatFileSplit(result.Value.(string))
		case "SOURCE":
			if err := genbankSourceParser(locusDepth)(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.Source = result.Value.(gt1.Organism)
		case "REFERENCE":
			if err := genbankReferenceParser(locusDepth)(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.References = append(metadata.References, result.Value.(gt1.Reference))
		case "COMMENT":
			if err := fieldBodyParser(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			metadata.Comment = result.Value.(string)
		}
	}
}
