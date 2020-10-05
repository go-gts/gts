package flags

import "fmt"

// Version represents the version of the command.
type Version struct {
	Major int
	Minor int
	Patch int
}

// String satisifers the fmt.Stringer interface.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
