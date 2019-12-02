package gt1

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ktnyt/ascii"
	"github.com/ktnyt/pars"
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
		panic(fmt.Sprintf("gt1 does not know how to format a qualifier of name `%s`", q.Name))
	}
}

// Format will format the qualifier.
func (q Qualifier) Format(prefix string) string {
	return prefix + strings.ReplaceAll(q.String(), "\n", "\n"+prefix)
}

// Qualifiers represents a collection of feature qualifiers.
type Qualifiers map[string][]string

// Get will return the qualifier values associated to the given name.
func (q Qualifiers) Get(key string) []string {
	if q == nil {
		return nil
	}
	if v, ok := q[key]; ok {
		return v
	}
	return nil
}

// Set will overwrite the qualifier values associated to the given name.
func (q Qualifiers) Set(name string, values ...string) { q[name] = values }

// Add will add a value to the qualifier associated to the given name.
func (q Qualifiers) Add(name, value string) { q[name] = append(q[name], value) }

// Del will delete the qualifier values associated to the given name.
func (q Qualifiers) Del(name string) { delete(q, name) }

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

// RegisterQuotedQualifier will register the given qualifier names as being a
// quoted qualifer (i.e. /name="value").
func RegisterQuotedQualifier(names ...string) {
	QuotedQualifierNames = append(QuotedQualifierNames, names...)
	sort.Strings(QuotedQualifierNames)
}

// RegisterLiteralQualifier will register the given qualifier names as being a
// literal qualifier (i.e. /name=value).
func RegisterLiteralQualifier(names ...string) {
	LiteralQualifierNames = append(LiteralQualifierNames, names...)
	sort.Strings(LiteralQualifierNames)
}

// RegisterToggleQualifier will register the given qualifier names as being a
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

// GetQualifierType will return the qualifier type of the given qualifier name.
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
	return pars.Seq(prefix+"/", pars.Word(ascii.IsSnake)).Child(1)
}

func quotedQualifierMap(prefix string) pars.Map {
	parser := pars.Delim(pars.Line, prefix)
	return func(result *pars.Result) error {
		state := pars.FromBytes(result.Token)
		if err := parser(state, result); err != nil {
			return err
		}
		ss := make([]string, len(result.Children))
		for i, child := range result.Children {
			ss[i] = string(child.Token)
		}
		result.SetValue(strings.Join(ss, "\n"))
		return nil
	}
}

func quotedQualifierParser(prefix string) pars.Parser {
	parser := pars.Seq('=', pars.Quoted('"')).Child(1)
	mapping := quotedQualifierMap(prefix)
	return parser.Map(mapping)
}

func chronobreak(parser pars.Parser) pars.Parser {
	return func(state *pars.State, result *pars.Result) error {
		state.Push()
		if err := parser(state, result); err != nil {
			state.Pop()
			return err
		}
		state.Pop()
		return nil
	}
}

// QualfierParser will attempt to match a single qualifier name-value pair.
func QualifierParser(prefix string) pars.Parser {
	wordParser := pars.Word(ascii.Not(ascii.IsSpace)).ToString()
	nameParser := qualifierNameParser(prefix)

	quotedParser := quotedQualifierParser(prefix)
	literalParser := pars.Seq('=', wordParser).Child(1)
	toggleParser := chronobreak(pars.Any('\n', pars.End)).Bind("")

	valueParsers := []pars.Parser{quotedParser, literalParser, toggleParser}

	return func(state *pars.State, result *pars.Result) error {
		if err := nameParser(state, result); err != nil {
			return nil
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
			}
		default:
			valueParser := valueParsers[qtype]
			if err := valueParser(state, result); err != nil {
				return err
			}
		}

		value := result.Value.(string)

		result.SetValue(Qualifier{name, value})
		return nil
	}
}
