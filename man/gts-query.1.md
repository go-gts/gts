# gts-query(1) -- query information from the given sequence

## SYNOPSIS

gts-query [--version] [-h | --help] [<args>] <input>

## DESCRIPTION

**gts-query** takes a single sequence input and reports various information
about its features. By default, it will output the sequence ID (or a unique
sequence number if there are no IDs available), a feature key, its location,
and any qualifiers that are common to all of the features present. A single
line represents a single feature entry.

This command is best utilized in combination with the gts-select(1) command.
Use gts-select(1) to narrow down the features to be extracted, and then apply
**gts-extract** to retrieve information. See the EXAMPLES section for more
insight. For a brief summary of a sequence, consider using gts-summary(1).

## OPTIONS

  * `<input>`:
    Input sequence file (may be omitted if standard input is provided). See
    gts-seqin(7) for a list of currently supported list of sequence formats.

  * `-d <delimiter>`, `--delimiter=<delimiter>`:
    String to insert between columns. The default delimiter is a tab `\t`
    character.

  * `--empty`:
    Allow missing qualifiers to be reported. Unlink GFFs, these columns will be
    completely empty.

  * `--no-header`:
    Do not print the header line.

  * `--no-key`:
    Do not report the feature key.

  * `--no-location`:
    Do not report the feature location.

  * `-n <name>`, `--name=<name>`:
    Qualifier name(s) to select. Multiple values may be set by repeatedly
    passing this option to the command. If set, only qualifiers that have the
    given name will be reported.

  * `-o <output>`, `--output=<output>`:
    Output table file (specifying `-` will force standard output).

  * `--source`:
    Include the source feature(s).

  * `-t <separator>`, `--separator=<separator>`:
    String to insert between qualifier values. The default separator is a comma
    `,` character. By default, the qualifier values will be reported in a CSV
    format. All commas and double quotes will be escaped, and all newline
    characters will be replaced with a whitespace.

## EXAMPLES

Report information of all CDS features:

    $ gts select CDS <input> | gts query

Report information of a CDS feature with `locus_tag` of `b0001`:

    $ gts select CDS/locus_tag=b0001 <input> | gts query

Report all of the `db_xref` qualifiers for every gene in the sequence:

    $ gts select gene | gts query -n db_xref
    $ gts select gene | gts query --name db_xref

## BUGS

**gts-query** currently has no known bugs.

## AUTHORS

**gts-query** is written and maintained by @AUTHOR@.

## SEE ALSO

gts(1), gts-select(1), gts-summary(1), gts-seqin(7), gts-seqout(7)