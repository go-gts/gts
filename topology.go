package gts

import (
	"fmt"
	"strings"
)

// Topology represents the sequence topology.
type Topology int

const (
	// Linear represents a linear sequence.
	Linear Topology = iota

	// Circular represents a circular sequence.
	Circular
)

// AsTopology converts a string to a Topology object.
func AsTopology(s string) (Topology, error) {
	switch strings.ToLower(s) {
	case "linear":
		return Linear, nil
	case "circular":
		return Circular, nil
	default:
		return Topology(-1), fmt.Errorf("unknown topology: %q", s)
	}
}

// String satisfies the fmt.Stringer interface.
func (t Topology) String() string {
	switch t {
	case Linear:
		return "linear"
	case Circular:
		return "circular"
	default:
		return ""
	}
}

type withTopology interface {
	WithTopology(t Topology) Sequence
}

// WithTopology creates a shallow copy of the given Sequence object and swaps
// the topology value with the given topology. If the sequence implements the
// `WithTopology(t Topoplogy) Sequence` method, it will be called instead.
func WithTopology(seq Sequence, t Topology) Sequence {
	switch v := seq.(type) {
	case withTopology:
		return v.WithTopology(t)
	default:
		return seq
	}
}
