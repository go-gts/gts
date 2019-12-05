package gts_test

import (
	"testing"

	"gopkg.in/ktnyt/assert.v1"
	"gopkg.in/ktnyt/gts.v0"
)

func TestComplement(t *testing.T) {
	original := gts.Seq("atgcATGCnN")
	complement := gts.Seq("tacgTACGnN")
	assert.Apply(t, assert.Equal(gts.Complement(original), complement))
}
