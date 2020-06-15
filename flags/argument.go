package flags

// Argument represents a value-usage pair.
type Argument struct {
	Value Value
	Usage string
}

// Arguments is a map of names and arguments.
type Arguments map[string]Argument

// Has tests if the given name is registered as an argument.
func (args Arguments) Has(name string) bool {
	_, ok := args[name]
	return ok
}
