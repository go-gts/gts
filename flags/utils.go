package flags

import "strings"

func shift(ss []string) (string, []string) {
	if len(ss) > 0 {
		return ss[0], ss[1:]
	}
	return "", nil
}

func sentencify(s string) string {
	if len(s) > 0 {
		s = strings.ToUpper(s[:1]) + s[1:]
		if s[len(s)-1] != '.' {
			s = s + "."
		}
	}
	return s
}
