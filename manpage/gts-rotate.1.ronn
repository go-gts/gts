# gts-rotate -- shift the coordinates of a circular sequence

## SYNOPSIS

gts-rotate [--version] [-h | --help] [<args>] <amount> <seqin>

## DESCRIPTION

**gts-rotate** takes a single sequence input and shifts the sequence with the
specified amount. For a positive shift, the start of the sequence will be moved
downstream relative to the strand represented by the sequence. For a negative
shift, the inverse is true. Because command options are recognized with a `-`
character, directly specifying a negative shift will result in erroneous
behavior. If a negative shift is required, either use the `-v` or `--backward`
option or insert a `--` after any options to force the command to interpret the
negative number as a literal value. See the EXAMPLES section for more insight.

## OPTIONS

  * `<amount>`:
    The amount to rotate the sequence by.

  * `<seqin>`:
    Input sequence file (may be omitted if standard input is provided).

  * `-F <format>`, `--format=<format>`:
    Output file format (defaults to same as input).

  * `-o <output>`, `--output=<output>`:
    Output sequence file (specifying `-` will force standard output).

  * `-v`, `--backward`:
    Rotate the sequence backwards (equivalent to a negative amount).

## EXAMPLES

Rotate a sequence 100 bases:

    $ gts rotate 100 <seqin>

Rotate a sequence -100 bases:

    $ gts rotate -v 100 <seqin>
    $ gts rotate --backward 100 <seqin>
    $ gts rotate -- 100 <seqin>

## BUGS

**gts-rotate** currently has no known bugs.

## AUTHORS

**gts-rotate** is written and maintained by Kotone Itaya.

## SEE ALSO

gts(1), gts-seqin(7)