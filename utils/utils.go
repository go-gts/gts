package utils

// Unpack a 2-element slice to its elements.
func Unpack(pair [2]int) (int, int) {
	return pair[0], pair[1]
}

// Min returns the smaller integer.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Compare two integers.
func Compare(a, b int) int {
	switch {
	case a < b:
		return -1
	case b < a:
		return 1
	default:
		return 0
	}
}
