package cache

import (
	"compress/flate"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"
)

// File represents a cache file.
type File struct {
	h  hash.Hash
	f  *os.File
	hd Header
	rd io.ReadCloser
	wr io.WriteCloser
}

// Open a cache file for reading.
func Open(path string, h hash.Hash, rsum, dsum []byte) (*File, error) {
	h.Reset()
	h.Write(append(rsum, dsum...))
	lsum := h.Sum(nil)
	name := filepath.Join(path, hex.EncodeToString(lsum))

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	size := h.Size()
	hd, err := ReadHeader(f, size)
	if err != nil {
		return nil, err
	}

	h.Reset()
	_, ret := io.Copy(h, f)
	bsum := h.Sum(nil)

	if err := hd.Validate(rsum, dsum, bsum); ret == nil {
		ret = err
	}

	if _, err := f.Seek(int64(size)*3, io.SeekStart); ret == nil {
		ret = err
	}

	rd := flate.NewReader(f)
	return &File{h, f, hd, rd, nil}, ret
}

// CreateLevel creates a new cache file with the given compression level.
func CreateLevel(path string, h hash.Hash, rsum, dsum []byte, level int) (*File, error) {
	h.Reset()
	h.Write(append(rsum, dsum...))
	lsum := h.Sum(nil)
	name := filepath.Join(path, hex.EncodeToString(lsum))

	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	_, ret := f.Write(make([]byte, h.Size()*3))

	hd := Header{rsum, dsum, nil}
	rd := flate.NewReader(f)
	wr, err := flate.NewWriter(f, level)
	if ret == nil {
		ret = err
	}
	return &File{h, f, hd, rd, wr}, ret
}

// Create a new cache file with the default compression level.
func Create(path string, h hash.Hash, rsum, dsum []byte) (*File, error) {
	return CreateLevel(path, h, rsum, dsum, flate.DefaultCompression)
}

// Name returns the name of the file.
func (f *File) Name() string {
	return f.f.Name()
}

// Read from the file through the flate.Reader.
func (f *File) Read(p []byte) (int, error) {
	return f.rd.Read(p)
}

// ReadOnly returns true if the file is read only.
func (f *File) ReadOnly() bool {
	return f.wr == nil
}

// Write to the file through the flate.Writer.
func (f *File) Write(p []byte) (int, error) {
	if f.wr == nil {
		return 0, io.EOF
	}
	return f.wr.Write(p)
}

// Close the files and write the body hash sum to the header.
func (f *File) Close() error {
	defer f.f.Close()
	defer f.rd.Close()

	if f.wr != nil {
		ret := f.wr.Close()

		if _, err := f.f.Seek(int64(f.h.Size())*3, io.SeekStart); ret == nil {
			ret = err
		}

		f.h.Reset()
		if _, err := io.Copy(f.h, f.f); ret == nil {
			ret = err
		}

		f.hd.BodySum = f.h.Sum(nil)
		if _, err := f.f.Seek(0, io.SeekStart); ret == nil {
			ret = err
		}

		if _, err := f.hd.WriteTo(f.f); ret == nil {
			ret = err
		}

		return ret
	}

	return nil
}
