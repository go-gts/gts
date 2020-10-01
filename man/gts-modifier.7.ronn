## gts-modifier(7) -- patterns for modifying sequence locations


## SYNOPSIS

^[(+|-)n]
$[(+|-)m]
^[(+|-)n]..$[(+|-)m]
^[(+|-)n]..^[(+|-)m]
$[(+|-)n]..$[(+|-)m]

## DESCRIPTION

**gts-modifier**s are patterns for modifying locations within a sequence. A
_modifier_ can take one of five forms: `^[(+|-)n]`, `$[[(+|-)m]]`,
`^[(+|-)n]..$[(+|-)m]`, `^[(+|-)n]..^[(+|-)m]`, or `$[(+|-)n]..$[(+|-)m]`.
A caret `^` character denotes the beginning of the location and a dollar `$`
character denotes the end of the location. The numbers following these
characters denote the offset of the position, where a negative number
represents the 5' region and a positive number represents the 3' region. The
first two forms of the _modifier_ will return a singular point location and the
latter three forms will return a modified range location. The positions and
offset values will be flipped for complement locations.

## EXAMPLES

Collapse the location to the start of the region:

    ^

Collapse the location to the end of the region:

    $

Leave the entire region as is:

    ^..$

Extend the region 20 bases upstream:

    ^-20..$

Focus the 20 bases around the end of the region:

    $-20..$+20

## SEE ALSO

gts(1), gts-extract(1), gts-locator(7)