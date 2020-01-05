package gts

import (
	"fmt"
	"io"
)

// Formatter can be formatted to a string or write to an io.Writer.
type Formatter interface {
	fmt.Stringer
	io.WriterTo
}
