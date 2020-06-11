package flags

// Value represents a command line argument value.
type Value interface {
	Set(value string) error
	String() string
}

// SliceValue represents a variable length command line argument value.
type SliceValue interface {
	Value
	Len() int
}
