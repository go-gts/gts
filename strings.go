package gts

import "strings"

// FlatFileSplit will split the string with the flatfile convention.
func FlatFileSplit(s string) []string {
	s = strings.TrimSuffix(s, ".")
	if len(s) > 0 {
		return strings.Split(s, "; ")
	}
	return []string{}
}
