# gts-reverse -- reverse order of the given sequence(s)

## SYNOPSIS

gts-reverse [--version] [-h | --help] [<args>] <input>

## DESCRIPTION

**gts-reverse** takes a single sequence input and reverses the sequence. Any
features present in the seqeuence will be relocated to match the reversed
location. This command _will not_ complement the sequence. To obtain the
complemented sequence, use **gts-complement(1)**.

## OPTIONS

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

## BUGS

**gts-reverse** currently has no known bugs.

## AUTHORS

**gts-reverse** is written and maintained by Kotone Itaya.

## SEE ALSO

gts(1), gts-complement(1), gts-seqin(7), gts-seqout(7)