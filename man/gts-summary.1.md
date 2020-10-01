# gts-summary(1) -- report a brief summary of the sequence(s)

## SYNOPSIS

gts-summary [--version] [-h | --help] [<args>] <input>

## DESCRIPTION

**gts-summary** takes a single sequence input and returns a brief summary of
its contents. By defalt, it will report the description, length, sequence
composition, feature counts, and qualifier counts. Use gts-query(1) to retrieve
more elaborate information of features.

## OPTIONS

  * `<input>`:
    Input sequence file (may be omitted if standard input is provided). See
    gts-seqin(7) for a list of currently supported list of sequence formats.

  * `--no-feature`:
    Suppress feature summary.

  * `--no-qualifier`:
    Suppress qualifier summary.

  * `-o <output>`, `--output=<output>`:
    Output file (specifying `-` will force standard output).

## BUGS

**gts-summary** currently has no known bugs.

## AUTHORS

**gts-summary** is written and maintained by @AUTHOR@.

## SEE ALSO

gts(1), gts-query(1), gts-seqin(7)