package gts

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	ascii "gopkg.in/ascii.v1"
	pars "gopkg.in/pars.v2"
)

// QualifierIO represents a single qualifier name-value pair.
type QualifierIO [2]string

// Unpack returns the name and value strings of the QualifierIO.
func (q QualifierIO) Unpack() (string, string) {
	return q[0], q[1]
}

// String satisfies the fmt.Stringer interface.
func (q QualifierIO) String() string {
	name, value := q.Unpack()
	switch GetQualifierType(name) {
	case QuotedQualifier:
		return fmt.Sprintf("/%s=\"%s\"", name, value)
	case LiteralQualifier:
		return fmt.Sprintf("/%s=%s", name, value)
	case ToggleQualifier:
		return "/" + name
	default:
		return fmt.Sprintf("/%s=\"%s\"", name, value)
	}
}

// Format creates a QualifierFormatter object for the qualifier with the given
// prefix.
func (q QualifierIO) Format(prefix string) QualifierFormatter {
	return QualifierFormatter{q, prefix}
}

// QualifierFormatter formats a QualifierIO object with the given prefix.
type QualifierFormatter struct {
	Qualifier QualifierIO
	Prefix    string
}

// String satisfies the fmt.Stringer interface.
func (qf QualifierFormatter) String() string {
	s := qf.Qualifier.String()
	return qf.Prefix + strings.ReplaceAll(s, "\n", "\n"+qf.Prefix)
}

// WriteTo satisfies the io.WriterTo interface.
func (qf QualifierFormatter) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, qf.String())
	return int64(n), err
}

// Names of qualifiers.
var (
	QuotedQualifierNames = []string{
		"allele", "altitude", "artificial_location", "bio_material",
		"bound_moiety", "cell_line", "cell_type", "chromosome",
		"clone", "clone_lib", "collected_by", "collection_date",
		"country", "cultivar", "culture_collection", "db_xref",
		"dev_stage", "EC_number", "ecotype", "exception",
		"experiment", "frequency", "function", "gap_type", "gene",
		"gene_synonym", "haplogroup", "haplotype", "host",
		"identified_by", "inference", "isolate", "isolation_source",
		"lab_host", "lat_lon", "linkage_evidence", "locus_tag", "map",
		"mating_type", "metagenome_source", "mobile_element_type",
		"mol_type", "ncRNA_class", "note", "old_locus_tag", "operon",
		"organelle", "organism", "PCR_conditions", "PCR_primers",
		"phenotype", "plasmid", "pop_variant", "product",
		"protein_id", "pseudogene", "recombination_class",
		"regulatory_class", "replace", "rpt_family", "rpt_unit_seq",
		"satellite", "segment", "serotype", "serovar", "sex",
		"specimen_voucher", "standard_name", "strain", "sub_clone",
		"submitter_seqid", "sub_species", "sub_strain", "tissue_lib",
		"tissue_type", "translation", "type_material", "variety",
	}

	LiteralQualifierNames = []string{
		"anticodon", "citation", "codon_start", "compare",
		"direction", "estimated_length", "mod_base", "number",
		"rpt_type", "rpt_unit_range", "tag_peptide", "transl_except",
		"transl_table",
	}

	ToggleQualifierNames = []string{
		"environmental_sample", "focus", "germline", "macronuclear",
		"partial", "proviral", "pseudo", "rearranged",
		"ribosomal_slippage", "transgenic", "trans_splicing",
	}
)

func init() {
	sort.Strings(QuotedQualifierNames)
	sort.Strings(LiteralQualifierNames)
	sort.Strings(ToggleQualifierNames)
}

// RegisterQuotedQualifier registers the given qualifier names as being a
// quoted qualifer (i.e. /name="value").
func RegisterQuotedQualifier(names ...string) {
	QuotedQualifierNames = append(QuotedQualifierNames, names...)
	sort.Strings(QuotedQualifierNames)
}

// RegisterLiteralQualifier registers the given qualifier names as being a
// literal qualifier (i.e. /name=value).
func RegisterLiteralQualifier(names ...string) {
	LiteralQualifierNames = append(LiteralQualifierNames, names...)
	sort.Strings(LiteralQualifierNames)
}

// RegisterToggleQualifier registers the given qualifier names as being a
// toggle qualifier (i.e. /name).
func RegisterToggleQualifier(names ...string) {
	ToggleQualifierNames = append(ToggleQualifierNames, names...)
	sort.Strings(ToggleQualifierNames)
}

func searchString(s string, ss []string) bool {
	if len(ss) == 0 {
		return false
	}
	n := len(ss) / 2
	l, m, r := ss[:n], ss[n], ss[n+1:]
	switch {
	case s < m:
		return searchString(s, l)
	case s > m:
		return searchString(s, r)
	default:
		return true
	}
}

// IsQuotedQualifier tests if the given qualifier name is a quoted qualifier.
func IsQuotedQualifier(name string) bool {
	return searchString(name, QuotedQualifierNames)
}

// IsLiteralQualifier tests if the given qualifier name is a literal qualifier.
func IsLiteralQualifier(name string) bool {
	return searchString(name, LiteralQualifierNames)
}

// IsToggleQualifier tests if the given qualifier name is a toggle qualifier.
func IsToggleQualifier(name string) bool {
	return searchString(name, ToggleQualifierNames)
}

// QualifierType represents the type of qualifier.
type QualifierType int

// Available qualifier types.
const (
	QuotedQualifier QualifierType = iota
	LiteralQualifier
	ToggleQualifier
	UnknownQualifier
)

// GetQualifierType returns the qualifier type of the given qualifier name.
func GetQualifierType(name string) QualifierType {
	switch {
	case IsQuotedQualifier(name):
		return QuotedQualifier
	case IsLiteralQualifier(name):
		return LiteralQualifier
	case IsToggleQualifier(name):
		return ToggleQualifier
	default:
		return UnknownQualifier
	}
}

func qualifierNameParser(prefix string) pars.Parser {
	p := []byte(prefix + "/")
	word := pars.Word(ascii.IsSnake)
	return func(state *pars.State, result *pars.Result) error {
		if err := state.Request(len(p)); err != nil {
			return err
		}
		if !bytes.Equal(state.Buffer(), p) {
			return pars.NewError(fmt.Sprintf("expected %q", prefix+"/"), state.Position())
		}
		state.Advance()
		return word(state, result)
	}
}

func quotedQualifierParser(prefix string) pars.Parser {
	quoted := pars.Quoted('"')
	p := append([]byte{'\n'}, []byte(prefix)...)
	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		c, err := pars.Next(state)
		if err != nil {
			state.Pop()
			return err
		}
		if c != '=' {
			state.Pop()
			return pars.NewError("expected `=`", state.Position())
		}
		state.Advance()
		if err := quoted(state, result); err != nil {
			state.Pop()
			return err
		}
		state.Drop()
		token := result.Token
		i := bytes.Index(token, p)
		for i >= 0 {
			n := copy(token[i+1:], token[i+len(p):])
			token = token[:i+1+n]
			i = bytes.Index(token, p)
		}
		result.SetToken(token)
		return nil
	}
}

func literalQualifierParser(prefix string) pars.Parser {
	literal := pars.Until(pars.Any("\n"+prefix+"/", pars.End))
	p := append([]byte{'\n'}, []byte(prefix)...)
	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		c, err := pars.Next(state)
		if err != nil {
			state.Pop()
			return err
		}
		if c != '=' {
			state.Pop()
			return pars.NewError("expected `=`", state.Position())
		}
		state.Advance()
		if err := literal(state, result); err != nil {
			state.Pop()
			return err
		}
		state.Drop()
		token := result.Token
		i := bytes.Index(token, p)
		for i >= 0 {
			n := copy(token[i+1:], token[i+len(p):])
			token = token[:i+1+n]
			i = bytes.Index(token, p)
		}
		result.SetToken(token)
		return nil
	}
}

// QualifierParser attempts to match a single qualifier name-value pair.
func QualifierParser(prefix string) pars.Parser {
	nameParser := qualifierNameParser(prefix)

	quotedParser := quotedQualifierParser(prefix)
	literalParser := literalQualifierParser(prefix)
	toggleParser := pars.Dry(pars.EOL)

	valueParsers := []pars.Parser{quotedParser, literalParser, toggleParser}

	return func(state *pars.State, result *pars.Result) error {
		if err := nameParser(state, result); err != nil {
			return err
		}
		name := string(result.Token)

		switch qtype := GetQualifierType(name); qtype {
		case UnknownQualifier:
			switch {
			case quotedParser(state, result) == nil:
				RegisterQuotedQualifier(name)
			case literalParser(state, result) == nil:
				RegisterLiteralQualifier(name)
			case toggleParser(state, result) == nil:
				RegisterToggleQualifier(name)
			default:
				return pars.NewError("unable to parse qualifier", state.Position())
			}
		default:
			if err := valueParsers[qtype](state, result); err != nil {
				return err
			}
		}

		value := string(result.Token)
		result.SetValue(QualifierIO{name, value})
		return nil
	}
}
