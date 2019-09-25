package gd_test

import (
	"testing"

	"github.com/ktnyt/gd"
	"github.com/stretchr/testify/require"
)

func TestSequence(t *testing.T) {
	s := "atgc"
	seq := gd.Seq(s)

	e := s
	t.Run("returns sequence", func(t *testing.T) {
		require.Equal(t, seq.String(), e)
	})

	e = "atatgcgc"
	t.Run("inserts sequence", func(t *testing.T) {
		seq.Insert(2, seq)
		require.Equal(t, seq.String(), e)
	})

	e = "atgcatgc"
	t.Run("replaces sequence", func(t *testing.T) {
		seq.Replace(2, gd.Seq("gcat"))
		require.Equal(t, seq.String(), e)
	})

	e = "atgc"
	t.Run("deletes sequence", func(t *testing.T) {
		seq.Delete(2, 4)
		require.Equal(t, seq.String(), e)
	})
}

func TestSequenceView(t *testing.T) {
	s := "atatgcgc"
	seq := gd.Seq(s)
	view := seq.View(2, 6)

	e := "atgc"
	t.Run("returns sequence", func(t *testing.T) {
		require.Equal(t, view.String(), e)
		require.Equal(t, seq.String(), "at"+e+"gc")
	})

	e = "atatgcgc"
	t.Run("inserts sequence", func(t *testing.T) {
		view.Insert(2, view)
		require.Equal(t, view.String(), e)
		require.Equal(t, seq.String(), "at"+e+"gc")
	})

	e = "atgcatgc"
	t.Run("replaces sequence", func(t *testing.T) {
		view.Replace(2, gd.Seq("gcat"))
		require.Equal(t, view.String(), e)
		require.Equal(t, seq.String(), "at"+e+"gc")
	})

	e = "atgc"
	t.Run("deletes sequence", func(t *testing.T) {
		view.Delete(2, 4)
		require.Equal(t, view.String(), e)
		require.Equal(t, seq.String(), "at"+e+"gc")
	})
}
