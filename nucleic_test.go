package gts_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gts"
)

func TestComplement(t *testing.T) {
	original := gts.Seq("atgcATGCnN")
	complement := gts.Seq("tacgTACGnN")
	assert.Apply(t, assert.Equal(gts.Complement(original), complement))
}
