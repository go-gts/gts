package gts

// Scanner provides a convenient interface for reading data.
type Scanner interface {
	Scan() bool
	Value() interface{}
	Err() error
}
