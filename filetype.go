package gts

import "path/filepath"

// FileType represents a file type.
type FileType int

// Available file types in GTS.
const (
	DefaultFile FileType = iota
	GenBankFlat
	JSONFile
	YAMLFile
	MsgpackFile
	UnknownFile
)

// Detect returns the FileType associated to extension of the given filename.
func Detect(filename string) FileType {
	ext := filepath.Ext(filename)
	if ext != "" {
		ext = ext[1:]
	}
	return ToFileType(ext)
}

// ToFileType converts the file type name string to a FileType
func ToFileType(name string) FileType {
	switch name {
	case "default":
		return DefaultFile
	case "gb", "genbank":
		return GenBankFlat
	case "json":
		return JSONFile
	case "yml", "yaml":
		return YAMLFile
	case "msgp", "msgpack":
		return MsgpackFile
	default:
		return UnknownFile
	}
}
