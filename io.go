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

// Scanner provides a convenient interface for reading data.
type Scanner interface {
	Scan() bool
	Value() interface{}
	Err() error
}
