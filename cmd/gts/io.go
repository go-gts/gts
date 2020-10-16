package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type attachment struct {
	r io.Reader
	w io.Writer
}

func (a *attachment) Read(p []byte) (int, error) {
	n, err := a.r.Read(p)
	if err != nil {
		return 0, err
	}
	return a.w.Write(p[:n])
}

func attach(w io.Writer, r io.Reader) *attachment {
	return &attachment{r, w}
}

type tuple [2]interface{}

func encodePayload(tt []tuple) []byte {
	p, err := json.Marshal(tt)
	if err != nil {
		panic(err)
	}
	return p
}

func gtsCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return dir, err
	}
	dir = filepath.Join(dir, "gts-cache")
	return dir, os.MkdirAll(dir, 0755)
}

func digest(h hash.Hash, p []byte, pp ...[]byte) error {
	h.Reset()
	for i := range pp {
		if _, err := h.Write(pp[i]); err != nil {
			return err
		}
	}
	copy(p, h.Sum(nil))
	return nil
}

type cacheHeader struct {
	isum []byte
	dsum []byte
	osum []byte
	bsum []byte
}

func readCacheHeader(f *os.File, h hash.Hash) (cacheHeader, error) {
	header := cacheHeader{}

	size := h.Size()
	p := make([]byte, size*3)
	n, err := f.Read(p)
	if err != nil {
		return header, err
	}
	if n != len(p) {
		return header, errors.New("not enough bytes in cache header")
	}

	i, j := size, size*2

	header.isum, header.dsum, header.bsum = p[:i], p[i:j], p[j:]
	header.osum = make([]byte, h.Size())
	if err := digest(h, header.osum, header.isum, header.dsum); err != nil {
		return header, nil
	}

	ohex := encodeToString(header.osum)
	name := filepath.Base(f.Name())
	if name != ohex {
		return header, fmt.Errorf("filename %q does not match hash %q", name, ohex)
	}

	h.Reset()
	if _, err := io.Copy(h, f); err != nil {
		return header, nil
	}

	bsum := h.Sum(nil)
	if !bytes.Equal(bsum, header.bsum) {
		out, exp := encodeToString(bsum), encodeToString(header.bsum)
		err := fmt.Errorf("body hash mismatch: got %q, expected %q", out, exp)
		return header, err
	}

	_, err = f.Seek(int64(size)*3, io.SeekStart)
	return header, err
}

type ioDelegate struct {
	h hash.Hash

	in  *os.File
	out *os.File
	tmp *os.File

	w io.Writer

	isum []byte
	dsum []byte
	osum []byte
	bsum []byte

	rm bool
}

func newIODelegate(h hash.Hash, inpath, outpath string) (*ioDelegate, error) {
	infile, outfile := os.Stdin, os.Stdout

	if inpath != "-" {
		f, err := os.Open(inpath)
		if err != nil {
			return nil, err
		}
		infile = f
	}

	if outpath != "-" {
		f, err := os.Create(outpath)
		if err != nil {
			return nil, err
		}
		outfile = f
	}

	isum := make([]byte, h.Size())
	dsum := make([]byte, h.Size())
	osum := make([]byte, h.Size())
	bsum := make([]byte, h.Size())

	return &ioDelegate{
		h,
		infile,
		outfile,
		nil,
		outfile,
		isum,
		dsum,
		osum,
		bsum,
		false,
	}, nil
}

func (d *ioDelegate) initializeCache(path string) error {
	if d.out == os.Stdout {
		f, err := os.Create(path)
		if err != nil {
			return err
		}

		if n, err := f.Write(d.isum); n != len(d.isum) || err != nil {
			f.Close()
			os.Remove(f.Name())
			return err
		}

		if n, err := f.Write(d.dsum); n != len(d.dsum) || err != nil {
			f.Close()
			os.Remove(f.Name())
			return err
		}

		if n, err := f.Write(d.bsum); n != len(d.bsum) || err != nil {
			f.Close()
			os.Remove(f.Name())
			return err
		}

		d.tmp = f
		d.w = io.MultiWriter(d.out, d.tmp, d.h)
		d.h.Reset()
	}

	return nil
}

func (d *ioDelegate) Cache(data []byte) (bool, error) {
	dir, err := gtsCacheDir()
	if err != nil {
		return false, nil
	}

	d.dsum = make([]byte, d.h.Size())
	if err := digest(d.h, d.dsum, data); err != nil {
		return false, nil
	}

	if d.in == os.Stdin {
		// Write to a temporary file to enable seeking.
		f, err := ioutil.TempFile("", "gts-tmp-*")
		if err != nil {
			return false, nil
		}

		d.in = f
		d.rm = true

		if _, err := io.Copy(d.in, os.Stdin); err != nil {
			d.Close()
			return false, err
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			d.Close()
			return false, err
		}
	}

	d.h.Reset()
	if _, err := io.Copy(d.h, d.in); err != nil {
		if _, err := d.in.Seek(0, io.SeekStart); err != nil {
			return false, err
		}
		return false, nil
	}

	if _, err := d.in.Seek(0, io.SeekStart); err != nil {
		return false, err
	}

	d.isum = d.h.Sum(nil)

	d.osum = make([]byte, d.h.Size())
	if err := digest(d.h, d.osum, d.isum, d.dsum); err != nil {
		return false, nil
	}

	name := encodeToString(d.osum)
	path := filepath.Join(dir, name)

	f, err := os.Open(path)
	if err != nil {
		// Cache file does not exist.
		d.initializeCache(path)
		return false, nil
	}

	if _, err := readCacheHeader(f, d.h); err != nil {
		// Cache file is corrupted.
		os.Remove(path)
		d.initializeCache(path)
		return false, nil
	}

	defer f.Close()

	if _, err := io.Copy(d.out, f); err != nil {
		return false, nil
	}

	if d.out != os.Stdout {
		os.Remove(f.Name())
	}

	return true, nil
}

func (d *ioDelegate) Read(p []byte) (int, error) {
	return d.in.Read(p)
}

func (d *ioDelegate) Write(p []byte) (int, error) {
	return d.w.Write(p)
}

func (d *ioDelegate) Close() error {
	if d.rm {
		defer os.Remove(d.in.Name())
	}

	defer d.in.Close()
	defer d.out.Close()

	if d.tmp != nil {
		d.bsum = d.h.Sum(nil)
		size := d.h.Size()
		if _, err := d.tmp.Seek(int64(size)*2, io.SeekStart); err != nil {
			return os.Remove(d.tmp.Name())
		}
		if n, err := d.tmp.Write(d.bsum); n != size || err != nil {
			return os.Remove(d.tmp.Name())
		}
		return d.tmp.Close()
	}

	return nil
}
