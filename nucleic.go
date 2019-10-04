package gt1

func complementBase(b byte) byte {
	switch b {
	case 'a':
		return 't'
	case 't':
		return 'a'
	case 'g':
		return 'c'
	case 'c':
		return 'g'
	case 'A':
		return 'T'
	case 'T':
		return 'A'
	case 'G':
		return 'C'
	case 'C':
		return 'G'
	default:
		return b
	}
}

func Complement(seq Sequence) Sequence {
	s := seq.Bytes()
	r := make([]byte, len(s))
	for i, c := range s {
		r[i] = complementBase(c)
	}
	return Seq(r)
}
