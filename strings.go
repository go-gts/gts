package gt1

import (
	"regexp"
	"strings"
)

func wrap(s string, indent int) string {
	if len(s) < 80 {
		return s
	}
	return s[:79] + "\n" + wrap(strings.Repeat(" ", indent)+s[79:], indent)
}

func wrapSpace(s string, indent int) string {
	if len(s) < 80 {
		return s
	}

	// Search for the last space before the line limit.
	i := 79
	for i >= 0 && s[i] != ' ' {
		i--
	}

	// If the line is not breakable, find the closest space.
	if i == 0 {
		i = 79
		for i < len(s) && s[i] != ' ' {
			i++
		}
	}

	// If there are no spaces remaining, just return the whole thing.
	if i == len(s) {
		return s
	}

	t := s[:i]
	r := strings.Repeat(" ", indent-1) + s[i:]

	return t + "\n" + wrapSpace(r, indent)
}

func removeIndent(s string) string {
	re, err := regexp.Compile("\n +")
	if err != nil {
		panic(err)
	}
	return re.ReplaceAllString(s, " ")
}

func flatfileSplit(s string) []string {
	s = removeIndent(s)
	s = strings.TrimSuffix(s, ".")
	if len(s) > 0 {
		return strings.Split(s, "; ")
	}
	return make([]string, 0)
}
