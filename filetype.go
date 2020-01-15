package gts

import "path/filepath"

// FileType represents a file type.
type FileType int

const (
	GenBankFlat FileType = iota
	GenBankPack
	UnknownFile
)

func GetFileType(filename string) FileType {
	switch filepath.Ext(filename) {
	case ".gb", ".genbank":
		return GenBankFlat
	case ".gbmp":
		return GenBankPack
	default:
		return UnknownFile
	}
}
