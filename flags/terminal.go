package flags

import isatty "github.com/mattn/go-isatty"

func IsTerminal(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
