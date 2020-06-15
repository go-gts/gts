package flags

func shift(ss []string) (string, []string) {
	if len(ss) > 0 {
		return ss[0], ss[1:]
	}
	return "", nil
}
