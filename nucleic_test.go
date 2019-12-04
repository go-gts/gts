package gt1_test

import (
	"testing"

	"github.com/ktnyt/assert"
	"github.com/ktnyt/gt1"
)

func TestComplement(t *testing.T) {
	original := gt1.Seq("atgcATGCnN")
	complement := gt1.Seq("tacgTACGnN")
	assert.Apply(t, assert.Equal(gt1.Complement(original), complement))
}
