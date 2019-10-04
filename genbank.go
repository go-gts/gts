package gt1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ktnyt/pars"
)

var genbankFieldDepth = 12
var genbankFeatureIndent = 5
var genbankFeatureDepth = 21

func formatLocusGenBank(r Record) string {
	name := r.LocusName
	length := strconv.Itoa(r.Length())
	pad1 := strings.Repeat(" ", 28-(len(name)+len(length)))
	molecule := r.Molecule
	pad2 := strings.Repeat(" ", 8-len(molecule))
	geometry := r.Topology
	pad3 := strings.Repeat(" ", 9-len(geometry))
	division := r.Division
	date := strings.ToUpper(r.Dates[0].Format("02-Jan-2006"))
	return "LOCUS       " + name + pad1 + length + " bp    " + molecule + pad2 + geometry + pad3 + division + " " + date
}

func formatSourceGenBank(source Organism) string {
	lines := make([]string, 0, 3)
	lines = append(lines, wrapSpace("SOURCE      "+source.Species, genbankFieldDepth))
	lines = append(lines, wrapSpace("  ORGANISM  "+source.Name, genbankFieldDepth))
	if source.Taxon != nil {
		lines = append(lines, wrapSpace("            "+strings.Join(source.Taxon, "; ")+".", genbankFieldDepth))
	}
	return strings.Join(lines, "\n")
}

func formatReferenceGenBank(reference Reference) string {
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf(
		"REFERENCE   %-2d (bases %d to %d)",
		reference.Number, reference.Start, reference.End,
	))

	if reference.Authors != "" {
		lines = append(lines, wrapSpace("  AUTHORS   "+reference.Authors, genbankFieldDepth))
	}

	if reference.Group != "" {
		lines = append(lines, wrapSpace("  CONSRTM   "+reference.Group, genbankFieldDepth))
	}

	if reference.Title != "" {
		lines = append(lines, wrapSpace("  TITLE     "+reference.Title, genbankFieldDepth))
	}

	if reference.Journal != "" {
		lines = append(lines, wrapSpace("  JOURNAL   "+reference.Journal, genbankFieldDepth))
	}

	if reference.Xref != nil {
		if v, ok := reference.Xref["PUBMED"]; ok {
			lines = append(lines, wrapSpace("   PUBMED   "+v, genbankFieldDepth))
		}
	}

	if reference.Comment != "" {
		lines = append(lines, wrapSpace("  REMARK    "+reference.Comment, genbankFieldDepth))
	}

	return strings.Join(lines, "\n")
}

func formatFeatureGenBank(feature Feature) string {
	lines := make([]string, 0)
	featureKey := strings.Repeat(" ", genbankFeatureIndent) + feature.Key + strings.Repeat(" ", genbankFeatureDepth-(genbankFeatureIndent+len(feature.Key)))
	lines = append(lines, featureKey+feature.Location.Format())
	for _, pair := range feature.Properties.Iter() {
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
		if key == "translation" {
			property = wrap(property, genbankFeatureDepth)
		} else {
			property = wrapSpace(property, genbankFeatureDepth)
		}
		lines = append(lines, property)
	}
	return strings.Join(lines, "\n")
}

func FormatGenBank(gb Record) string {
	lines := make([]string, 0)
	lines = append(lines, formatLocusGenBank(gb))
	lines = append(lines, wrapSpace("DEFINITION  "+gb.Definition, genbankFieldDepth))
	lines = append(lines, "ACCESSION   "+gb.Accessions[0])
	lines = append(lines, "VERSION     "+gb.Version)
	if gb.DBLink.Len() > 0 {
		for i, pair := range gb.DBLink.Iter() {
			dblink := strings.Repeat(" ", genbankFieldDepth)
			if i == 0 {
				dblink = "DBLINK" + strings.Repeat(" ", genbankFieldDepth-6)
			}
			dblink += fmt.Sprintf("%s: %s", pair.Key, pair.Value)
			lines = append(lines, dblink)
		}
	}
	lines = append(lines, wrapSpace("KEYWORDS    "+strings.Join(gb.Keywords, "; ")+".", genbankFieldDepth))
	lines = append(lines, formatSourceGenBank(gb.Source))
	for _, reference := range gb.References {
		lines = append(lines, formatReferenceGenBank(reference))
	}
	if len(gb.Comment) > 0 {
		lines = append(lines, wrapSpace("COMMENT     "+gb.Comment, genbankFieldDepth))
	}
	lines = append(lines, "FEATURES             Location/Qualifiers")
	for _, feature := range gb.Features {
		lines = append(lines, formatFeatureGenBank(feature))
	}
	lines = append(lines, "ORIGIN      ")
	for i := 0; i < gb.Length(); i += 60 {
		seq := make([]string, 0, 6)
		for j := 0; j < 60 && i+j < gb.Length(); j += 10 {
			k := i + j + 10
			if i+j+10 > gb.Length() {
				k = gb.Length()
			}
			seq = append(seq, gb.Slice(i+j, k).String())
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
		dblink := NewPairList()
		for {
			if err := genbankDBLinkEntryParser(state, result); err != nil {
				return pars.NewTraceError("GenBank DBLink", err)
			}

			dblink.Set(result.Children[0].Value.(string), result.Children[1].Value.(string))

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
		source := Organism{Species: result.Value.(string)}

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
		source.Taxon = flatfileSplit(result.Value.(string))

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

		reference := Reference{
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

type genbankFeatureName struct {
	Indent int
	Value  string
	Depth  int
}

var genbankFeatureNameParser = pars.Seq(
	pars.Many(' '),
	pars.SnakeWord.Map(pars.CatByte),
	pars.Many(' '),
).Map(func(result *pars.Result) error {
	indent := len(result.Children[0].Children)
	value := result.Children[1].Value.(string)
	depth := indent + len(value) + len(result.Children[2].Children)
	result.Value = genbankFeatureName{Indent: indent, Value: value, Depth: depth}
	result.Children = nil
	return nil
})

func genbankFeatureBodyParser(indent, depth int) pars.Parser {
	depthString := "\n" + strings.Repeat(" ", depth)

	return func(state *pars.State, result *pars.Result) error {
		// First line must be a range.
		if err := locatableParser(state, result); err != nil {
			return err
		}
		location := result.Value.(Location)
		pars.Try('\n')(state, result)

		pairs := make([]Pair, 0)

		for {
			// Count the leading spaces.
			state.Mark()

			count := 0

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

			// End of feature so return.
			if count <= indent {
				state.Jump()
				result.Value = Feature{
					Location:   location,
					Properties: NewPairListFromPairs(pairs),
				}
				result.Children = nil
				return nil
			}

			if count != depth {
				state.Jump()
				return pars.NewMismatchError("GeBank Feature Field", []byte("matching depth"), state.Position)
			}

			// Remaining fields must be preceded by a /.
			if state.Buffer[state.Index] != '/' {
				state.Jump()
				return pars.NewMismatchError("GenBank Feature Field", []byte{'/'}, state.Position)
			}

			if err := state.Want(1); err != nil {
				state.Jump()
				return err
			}
			state.Advance(1)

			// Match the name of the feature field.
			if err := pars.Until(pars.Any('=', '\n'))(state, result); err != nil {
				state.Jump()
				return err
			}
			name := result.Value.(string)

			if state.Buffer[state.Index] == '\n' {
				pars.Try('\n')(state, result)
				pairs = append(pairs, Pair{Key: name, Value: ""})
				state.Unmark()
				continue
			}

			// Next byte is guaranteed to be = by Until.
			state.Advance(1)

			// In most cases the feature property values are quoted.
			if err := pars.Quoted('"')(state, result); err != nil {
				// Otherwise just get the rest of the line.
				if err := pars.Line(state, result); err != nil {
					state.Jump()
					return err
				}
			}
			value := result.Value.(string)

			// Completely remove indents for translations.
			if name == "translation" {
				value = strings.Replace(value, depthString, "", -1)
			} else {
				value = strings.Replace(value, depthString, " ", -1)
			}

			// Remove the newline.
			pars.Try('\n')(state, result)

			pairs = append(pairs, Pair{Key: name, Value: value})

			state.Unmark()
		}
	}
}

func genbankFeatureParser(state *pars.State, result *pars.Result) error {
	// Discard the Location/Qualifiers line.
	if err := pars.Line(state, result); err != nil {
		return pars.NewTraceError("GenBank Feature", err)
	}

	// The first feature must be a source.
	if err := genbankFeatureNameParser(state, result); err != nil {
		return pars.NewTraceError("GenBank Feature", err)
	}

	featureName := result.Value.(genbankFeatureName)

	if featureName.Value != "source" {
		return pars.NewMismatchError("GenBank Feature", []byte("source"), state.Position)
	}

	features := make([]Feature, 1)

	sourceIndent := featureName.Indent
	sourceDepth := featureName.Depth

	bodyParser := genbankFeatureBodyParser(sourceIndent, sourceDepth)

	// Process the source feature body.
	if err := bodyParser(state, result); err != nil {
		return pars.NewTraceError("GenBank Feature", err)
	}

	features[0] = result.Value.(Feature)
	features[0].Key = featureName.Value

	// Continually process feature properties while indented.
	for state.Buffer[state.Index] == ' ' {
		if err := genbankFeatureNameParser(state, result); err != nil {
			return pars.NewTraceError("GenBank Feature", err)
		}
		featureName = result.Value.(genbankFeatureName)

		if err := bodyParser(state, result); err != nil {
			return pars.NewTraceError("GenBank Feature", err)
		}
		feature := result.Value.(Feature)
		feature.Key = featureName.Value
		features = append(features, feature)
	}

	result.Value = features
	result.Children = nil
	return nil
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

	gb := Record{
		LocusName: result.Children[0].Value.(string),
		Molecule:  result.Children[2].Value.(string),
		Topology:  result.Children[3].Value.(string),
		Division:  result.Children[4].Value.(string),
		Dates:     []time.Time{result.Children[5].Value.(time.Time)},
	}

	pars.Try('\n')(state, result)

	// Continually process fields.
	for {
		// End of entry.
		if err := pars.AsParser("//")(state, result); err == nil {
			pars.Try('\n')(state, result)
			result.Value = gb
			result.Children = nil
			return nil
		}

		if err := genbankFieldNameParser(state, result); err != nil {
			return pars.NewTraceError("GenBank", err)
		}
		fieldName = result.Value.(genbankFieldName)

		// FEATURES and ORIGIN do not fit the field conventions.
		if fieldName.Value == "FEATURES" {
			if err := genbankFeatureParser(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.Features = result.Value.([]Feature)
			continue
		}

		if fieldName.Value == "ORIGIN" {
			pars.Try('\n')(state, result)
			origin := make([]byte, 0, length)

			for state.Buffer[state.Index] == ' ' {
				if err := genbankOriginLineParser(state, result); err != nil {
					return pars.NewTraceError("GenBank", err)
				}
				origin = append(origin, []byte(result.Value.(string))...)
				state.Clear()
			}
			gb.s = origin
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
			gb.Definition = result.Value.(string)
		case "ACCESSION":
			if err := pars.Line(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.Accessions = []string{result.Value.(string)}
		case "VERSION":
			if err := pars.Line(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.Version = result.Value.(string)
		case "DBLINK":
			if err := genbankDBLinkParser(locusDepth)(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.DBLink = result.Value.(PairList)
		case "KEYWORDS":
			if err := pars.Line(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.Keywords = flatfileSplit(result.Value.(string))
		case "SOURCE":
			if err := genbankSourceParser(locusDepth)(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.Source = result.Value.(Organism)
		case "REFERENCE":
			if err := genbankReferenceParser(locusDepth)(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.References = append(gb.References, result.Value.(Reference))
		case "COMMENT":
			if err := fieldBodyParser(state, result); err != nil {
				return pars.NewTraceError("GenBank", err)
			}
			gb.Comment = result.Value.(string)
		}
	}
}
