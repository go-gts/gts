# gts-select -- select features using the given feature selector(s)

## SYNOPSIS

gts-select [--version] [-h | --help] [<args>] <selector> <input>

## DESCRIPTION

**gts-select** takes a _selector_ and a single sequence input, and selects the
features which satisfy the _selector_ criteria. A _selector_ takes the form
`[feature_key][/[qualifier1][=regexp1]][/[qualifier2][=regexp2]]...`. See
gts-selector(7) for more details.

**gts-select** serves as a central command, allowing the user to filter out
features for use in other commands like gts-extract(1) and gts-query(1). See
the EXAMPLES section for more insight.

## OPTIONS

  * `<selector>`:
    Feature selector
    (syntax: [feature_key][/[qualifier1][=regexp1]][/[qualifier2][=regexp2]]...).
    See gts-selector(7) for more details.

  * `<input>`:
    Input sequence file (may be omitted if standard input is provided). See
    gts-seqin(7) for a list of currently supported list of sequence formats.

  * `-F <format>`, `--format=<format>`:
    Output file format (defaults to same as input). See gts-seqout(7) for a
    list of currently supported list of sequence formats. The format specified
    with this option will override the file type detection from the output
    filename.

  * `-o <output>`, `--output=<output>`:
    Output sequence file (specifying `-` will force standard output). The
    output file format will be automatically detected from the filename if none
    is specified with the `-F` or `--format` option.

  * `-v`, `--invert-match`:
    Select features that do not match the given criteria.

## EXAMPLES

Select all of the CDS features:

    $ gts select CDS <input>

Select all features with `locus_tag` of `b0001`:

    $ gts select /locus_tag=b0001 <input>

Select all features with the qualifier `translation`:

    $ gts select /translation <input>

Select all features with a qualifier value matching `recombinase`

    $ gts select /=recombinase <input>

## BUGS

**gts-select** currently has no known bugs.

## AUTHORS

**gts-select** is written and maintained by Kotone Itaya.

## SEE ALSO

gts(1), gts-query(1), gts-locator(7), gts-selector(7), gts-seqin(7),
gts-seqout(7)