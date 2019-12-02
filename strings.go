package gt1

import "strings"

// WrapByte will wrap the string at the last given byte located before the
// given width limit. If there is not such byte available, the string will be
// wrapped at the first occurence of the said byte. If the given byte does not
// exist in the string, the entire string will be returned.
func WrapByte(s string, width int, c byte) string {
	i := strings.IndexByte(s, '\n')
	if i >= 0 {
		return WrapByte(s[:i], width, c) + "\n" + WrapByte(s[i+1:], width, c)
	}

	if len(s) < width {
		return s
	}

	i = strings.LastIndexByte(s[:width], c)
	if i >= 0 {
		return s[:i] + "\n" + WrapByte(s[i+1:], width, c)
	}

	i = strings.IndexByte(s, c)
	if i >= 0 {
		return s[:i] + "\n" + WrapByte(s[i+1:], width, c)
	}

	return s
}

// Wrap will wrap the string at the given width.
func Wrap(s string, width int) string {
	i := strings.IndexByte(s, '\n')
	if i >= 0 {
		return Wrap(s[:i], width) + "\n" + Wrap(s[i+1:], width)
	}

	if len(s) < width {
		return s
	}

	return s[:width] + "\n" + Wrap(s[width:], width)
}
