package gts

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	ascii "gopkg.in/ktnyt/ascii.v1"
	pars "gopkg.in/ktnyt/pars.v2"
)

// Qualifier represents a single qualifier name-value pair.
type Qualifier struct {
	Name  string
	Value string
}

// String satisfies the fmt.Stringer interface.
func (q Qualifier) String() string {
	switch GetQualifierType(q.Name) {
	case QuotedQualifier:
		return fmt.Sprintf("/%s=\"%s\"", q.Name, q.Value)
	case LiteralQualifier:
		return fmt.Sprintf("/%s=%s", q.Name, q.Value)
	case ToggleQualifier:
		return "/" + q.Name
	default:
		panic(fmt.Sprintf("gts does not know how to format a qualifier of name `%s`", q.Name))
	}
}

// Format creates a QualifierFormatter object for the qualifier with the given
// prefix.
func (q Qualifier) Format(prefix string) QualifierFormatter {
	return QualifierFormatter{q, prefix}
}

// QualifierFormatter formats a Qualifier object with the given prefix.
type QualifierFormatter struct {
	Qualifier Qualifier
	Prefix    string
}

// String satisfies the fmt.Stringer interface.
func (qf QualifierFormatter) String() string {
	s := qf.Qualifier.String()
	return qf.Prefix + strings.ReplaceAll(s, "\n", "\n"+qf.Prefix)
}

// WriteTo satisfies the io.WriterTo interface.
func (qf QualifierFormatter) WriteTo(w io.Writer) (int, error) {
	return w.Write([]byte(qf.String()))
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
		s := string(result.Token)
		result.SetValue(strings.ReplaceAll(s, "\n"+prefix, "\n"))
		return nil
	}
}

func literalQualifierParser(prefix string) pars.Parser {
	literal := pars.Until(pars.Any("\n"+prefix+"/", pars.End))
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
		s := string(result.Token)
		result.SetValue(strings.ReplaceAll(s, "\n"+prefix, "\n"))
		return nil
	}
}

// QualfierParser attempts to match a single qualifier name-value pair.
func QualifierParser(prefix string) pars.Parser {
	nameParser := qualifierNameParser(prefix)

	quotedParser := quotedQualifierParser(prefix)
	literalParser := literalQualifierParser(prefix)
	toggleParser := pars.Dry(pars.EOL).Bind("")

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

		value := result.Value.(string)
		result.SetValue(Qualifier{name, value})
		return nil
	}
}