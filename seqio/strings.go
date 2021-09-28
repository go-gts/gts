package seqio

import (
	"strings"
)

func multiLineString(ss ...string) string {
	return strings.Join(ss, "\n")
}

// AddPrefix adds the given prefix after each newline.
func AddPrefix(s, prefix string) string {
	return strings.ReplaceAll(s, "\n", "\n"+prefix)
}
