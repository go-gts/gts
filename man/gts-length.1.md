# gts-length -- report the length of the sequence(s)

## SYNOPSIS

gts-length [--version] [-h | --help] [<args>] <input>

## DESCRIPTION

**gts-length** takes a single sequence input and prints the length of each
sequence in the given sequence file.

## OPTIONS

  * `<input>`:
    Input sequence file (may be omitted if standard input is provided). See
    gts-seqin(7) for a list of currently supported list of sequence formats.

  * `-o <output>`, `--output=<output>`:
    Output file (specifying `-` will force standard output).

## BUGS

**gts-length** currently has no known bugs.

## AUTHORS

**gts-length** is written and maintained by Kotone Itaya.

## SEE ALSO

gts(1), gts-seqin(7)