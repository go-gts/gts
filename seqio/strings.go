package seqio

import (
	"regexp"
	"strings"
)

func RemoveIndent(s string) string {
	re, err := regexp.Compile("\n +")
	if err != nil {
		panic(err)
	}
	return re.ReplaceAllString(s, " ")
}

func RemoveNewline(s string) string {
	return strings.Join(strings.Split(s, "\n"), "")
}

func FlatFileSplit(s string) []string {
	s = RemoveIndent(s)
	s = strings.TrimSuffix(s, ".")
	if len(s) > 0 {
		return strings.Split(s, "; ")
	}
	return make([]string, 0)
}

func Wrap(width int, prefix string) func(string) string {
	limit := width - 1

	return func(s string) string {
		if len(s) < width {
			return s
		}

		head, tail := s[:limit], s[limit:]
		return head + "\n" + Wrap(width, prefix)(prefix+tail)
	}
}

func WrapAt(at byte, width int, prefix string) func(string) string {
	limit := width - 1

	return func(s string) string {
		if len(s) < width {
			return s
		}

		pivot := strings.LastIndexByte(s[:limit], at)
		if pivot < len(prefix) {
			pivot = strings.IndexByte(s[limit:], at) + limit
		}

		head, tail := s[:pivot], s[pivot+1:]
		return head + "\n" + WrapAt(at, width, prefix)(prefix+tail)
	}
}
