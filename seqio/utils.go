package seqio

func dig(err error) error {
	if v, ok := err.(interface{ Unwrap() error }); ok {
		return dig(v.Unwrap())
	}
	return err
}
