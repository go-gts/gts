package gt1

import "strings"

func smartWrapImpl(ss []string, s string, width int) []string {
	if len(s) < width {
		return append(ss, s)
	}
	i := strings.LastIndexByte(s[:width], ' ')
	if i < 0 {
		i = width + strings.IndexByte(s[width:], ' ')
		if i < width {
			return append(ss, s)
		}
	}
	ss = append(ss, s[:i])
	return smartWrapImpl(ss, s[i+1:], width)
}

func smartWrap(s string, width int) []string {
	return smartWrapImpl([]string{}, s, width)
}

func wrapImpl(ss []string, s string, width int) []string {
	if len(s) < width {
		return append(ss, s)
	}
	ss = append(ss, s[:width])
	return wrapImpl(ss, s, width)
}

func wrap(s string, width int) []string {
	return wrapImpl([]string{}, s, width)
}
