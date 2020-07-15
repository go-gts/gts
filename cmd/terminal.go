package cmd

import "github.com/mattn/go-isatty"

// IsTerminal tests if the file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
