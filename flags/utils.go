package flags

import (
	"bufio"
	"io"
	"os"
	"strings"
)

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

func osAppend(filename string) (*os.File, error) {
	flag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	return os.OpenFile(filename, flag, 0644)
}

func fileAppend(filename, s string) error {
	f, err := osAppend(filename)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	if _, err := io.WriteString(w, s); err != nil {
		return err
	}
	return w.Flush()
}

func alignLines(s string, c byte) string {
	lines := strings.Split(s, "\n")
	indices := make([]int, len(lines))
	max := 0
	for i, line := range lines {
		index := strings.IndexByte(line, c)
		indices[i] = index
		if max < index {
			max = index
		}
	}
	for i, line := range lines {
		pad := ""
		if index := indices[i]; index > 0 {
			diff := max - index
			for j := 0; j < diff; j++ {
				pad += " "
			}
		}
		lines[i] = strings.Replace(line, string([]byte{c}), pad, 1)
	}
	return strings.Join(lines, "\n")
}

func touch(s string) error {
	f, err := os.Create(s)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
