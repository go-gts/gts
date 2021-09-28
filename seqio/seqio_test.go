package seqio

import (
	"math/rand"

	"github.com/go-gts/gts"
)

func init() {
	gts.SetLogLevel(gts.SILENT)
	rand.Seed(0)
}

func StringWithCharset(charset string, length int) string {
	runeset := []rune(charset)
	rr := make([]rune, length)
	for i := range rr {
		rr[i] = runeset[rand.Intn(len(runeset))]
	}
	return string(rr)
}
