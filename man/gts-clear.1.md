# gts-clear(1) -- remove all features from the sequence (excluding source features)

## SYNOPSIS

gts-clear [--version] [-h | --help] [<args>] <input>

## DESCRIPTION

**gts-clear** takes a single sequence file input and strips off all features
except for the `source` features which are mandatory in GenBank. This command
is equivalent to running `gts select source <input>`.

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

**gts-clear** currently has no known bugs.

## AUTHORS

**gts-clear** is written and maintained by Kotone Itaya.

## SEE ALSO

gts(1), gts-seqin(7), gts-seqout(7)