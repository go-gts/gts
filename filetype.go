package gts

import "path/filepath"

// FileType represents a file type.
type FileType int

// Available file types in GTS.
const (
	DefaultFile FileType = iota
	UnknownFile
	JSONFile
	YAMLFile
	MsgpackFile
	FastaFile
	FastqFile
	GenBankFile
	EMBLFile
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
	case "", "default":
		return DefaultFile
	case "json":
		return JSONFile
	case "yml", "yaml":
		return YAMLFile
	case "msgp", "msgpack":
		return MsgpackFile
	case "fasta":
		return FastaFile
	case "fastq":
		return FastqFile
	case "gb", "genbank":
		return GenBankFile
	case "emb", "embl":
		return EMBLFile
	default:
		return UnknownFile
	}
}
