## gts-selector(7) -- patterns to select sequence features

## SYNOPSIS

[feature_key][/[qualifier1][=regexp1]][/[qualifier2][=regexp2]]...

## DESCRIPTION

**gts-selector**s are patterns for selecting sequence features that match the
given _selector_. A _selector_ consists of a single feature key and/or multiple
qualifier matchers. A feature key must currently be a perfect match
(case sensitive) and if omitted all feature keys will match. A qualifier
matcher has two parts: a qualifier name and a regular expression delimited by
the `=` sign. The qualifier name must currently be a perfect match (case
sensitive) and if omitted all qualifier names will match. The regular
expression will be tested against the contents of the qualifier value. If
omitted, any features that has the qualifier with the given qualifier name will
match.

## EXAMPLES

Select all `gene` features:

    gene

Select all `CDS` features that produce a DNA-binding `product`:

    CDS/product=DNA-binding

Select all features with `locus_tag` of `b0001`:

    /locus_tag=b0001

Select all features with the qualifier `translation`:

    /translation

Select all features with a qualifier value matching `recombinase`

    /=recombinase

## SEE ALSO

gts(1), gts-select(1), gts-locator(7)