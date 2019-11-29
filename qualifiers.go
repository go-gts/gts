package gt1

import (
	"fmt"
	"strings"

	"github.com/ktnyt/ascii"
	"github.com/ktnyt/pars"
)

type qualifierValueType int

const (
	quotedQualifierValue qualifierValueType = iota
	translQualifierValue
	literalQualifierValue
	toggleQualifierValue
	unknownQualifierValue
)

func getQualifierValueType(name string) qualifierValueType {
	switch name {
	case "allele", "altitude", "artificial_location", "bio_material",
		"bound_moiety", "cell_line", "cell_type", "chromosome", "clone",
		"clone_lib", "collected_by", "collection_date", "country",
		"cultivar", "culture_collection", "db_xref", "dev_stage",
		"EC_number", "ecotype", "exception", "experiment", "frequency",
		"function", "gap_type", "gene", "gene_synonym", "haplogroup",
		"haplotype", "host", "identified_by", "inference", "isolate",
		"isolation_source", "lab_host", "lat_lon", "linkage_evidence",
		"locus_tag", "map", "mating_type", "metagenome_source",
		"mobile_element_type", "mol_type", "ncRNA_class", "note",
		"old_locus_tag", "operon", "organelle", "organism",
		"PCR_conditions", "PCR_primers", "phenotype", "plasmid",
		"pop_variant", "product", "protein_id", "pseudogene",
		"recombination_class", "regulatory_class", "replace",
		"rpt_family", "rpt_unit_seq", "satellite", "segment",
		"serotype", "serovar", "sex", "specimen_voucher",
		"standard_name", "strain", "sub_clone", "submitter_seqid",
		"sub_species", "sub_strain", "tissue_lib", "tissue_type",
		"type_material", "variety":
		return quotedQualifierValue
	case "anticodon", "citation", "codon_start", "compare", "direction",
		"estimated_length", "mod_base", "number", "rpt_type", "rpt_unit_range",
		"tag_peptide", "transl_except", "transl_table":
		return literalQualifierValue
	case "environmental_sample", "focus", "germline", "macronuclear",
		"partial", "proviral", "pseudo", "rearranged", "ribosomal_slippage",
		"transgenic", "trans_splicing":
		return toggleQualifierValue
	case "translation":
		return translQualifierValue
	default:
		return unknownQualifierValue
	}
}

// Qualifier represents a single qualifier name-value pair.
type Qualifier struct {
	Name  string
	Value string
}

// Format will format the qualifier.
func (q Qualifier) Format(indent, width int) string {
	var ss []string
	switch getQualifierValueType(q.Name) {
	case quotedQualifierValue:
		ss = smartWrap(fmt.Sprintf("/%s=\"%s\"", q.Name, q.Value), width-indent)
	case translQualifierValue:
		ss = wrap(fmt.Sprintf("/%s=\"%s\"", q.Name, q.Value), width-indent)
	case literalQualifierValue:
		ss = smartWrap(fmt.Sprintf("/%s=%s", q.Name, q.Value), width-indent)
	case toggleQualifierValue:
		ss = []string{fmt.Sprintf("/%s", q.Name)}
	default:
		panic(fmt.Sprintf("unknown qualifier name `%s`", q.Name))
	}
	for i := range ss {
		ss[i] = spaces(indent) + ss[i]
	}
	return strings.Join(ss, "\n")
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

func spaces(indent int) string { return strings.Repeat(" ", indent) }

func qualifierNameParser(indent int) pars.Parser {
	return pars.Seq(spaces(indent)+"/", pars.Word(ascii.IsSnake)).Child(1)
}

func quotedQualifierMap(indent int, sep string) pars.Map {
	parser := pars.Delim(pars.Line, spaces(indent))
	return func(result *pars.Result) error {
		state := pars.FromBytes(result.Token)
		if err := parser(state, result); err != nil {
			return err
		}
		ss := make([]string, len(result.Children))
		for i, child := range result.Children {
			ss[i] = string(child.Token)
		}
		result.SetValue(strings.Join(ss, sep))
		return nil
	}
}

func quotedQualifierParser(indent int, sep string) pars.Parser {
	parser := pars.Seq('=', pars.Quoted('"')).Child(1)
	mapping := quotedQualifierMap(indent, sep)
	return parser.Map(mapping)
}

func literalQualifierParser(indent int) pars.Parser {
	return pars.Seq('=', pars.Line).Child(1).Map(pars.ToString)
}

// QualfierParser will attempt to match a single qualifier name-value pair.
func QualifierParser(indent int) pars.Parser {
	wordParser := pars.Word(ascii.Not(ascii.IsSpace)).Map(pars.ToString)
	nameParser := qualifierNameParser(indent)
	valueParsers := []pars.Parser{
		quotedQualifierParser(indent, " "),
		quotedQualifierParser(indent, ""),
		pars.Seq('=', wordParser).Child(1),
		pars.AsParser(pars.Epsilon).Bind(""),
	}

	return func(state *pars.State, result *pars.Result) error {
		if err := nameParser(state, result); err != nil {
			return nil
		}
		name := string(result.Token)

		valueType := getQualifierValueType(name)
		if valueType == unknownQualifierValue {
			panic(fmt.Sprintf("unknown qualifier name `%s`", name))
		}

		valueParser := valueParsers[valueType]
		if err := valueParser(state, result); err != nil {
			return nil
		}
		value := result.Value.(string)

		result.SetValue(Qualifier{name, value})
		return nil
	}
}
