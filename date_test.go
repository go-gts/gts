package gts

import (
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	now := time.Now()
	in := FromTime(now)
	out := FromTime(in.ToTime())
	equals(t, in, out)
}
