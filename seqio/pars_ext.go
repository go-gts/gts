package seqio

import (
	"github.com/go-ascii/ascii"
)

const spaceByte = byte(' ')

var isBaseCharacter = ascii.Range(33, 126)
