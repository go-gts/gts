package gt1

func insertBytes(s []byte, pos int, vs []byte) []byte {
	r := make([]byte, len(s)+len(vs))
	copy(r[:pos], s[:pos])
	copy(r[pos:], vs)
	copy(r[pos+len(vs):], s[pos:])
	return r
}

func deleteBytes(s []byte, pos, cnt int) []byte {
	r := make([]byte, len(s)-cnt)
	copy(r[:pos], s[:pos])
	copy(r[pos:], s[pos+cnt:])
	return r
}

func replaceBytes(s []byte, pos int, vs []byte) []byte {
	r := make([]byte, len(s))
	copy(r[:pos], s[:pos])
	copy(r[pos:], vs)
	copy(r[pos+len(vs):], s[pos+len(vs):])
	return r
}
