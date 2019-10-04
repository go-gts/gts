package gt1

import "github.com/ktnyt/pars"

func notByte(t byte) pars.ByteFilter {
	return func(b byte) bool {
		return b != t
	}
}

func notFilter(f pars.ByteFilter) pars.ByteFilter {
	return func(b byte) bool {
		return !f(b)
	}
}
