package main

import (
	"compress/flate"
	"encoding/json"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-gts/gts/cmd/cache"
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

type ioDelegate struct {
	infile  *os.File
	outfile *os.File
	cache   *cache.File
	tmpin   bool
}

func newIODelegate(inpath, outpath string) (*ioDelegate, error) {
	input, output := os.Stdin, os.Stdout
	var err error

	if inpath != "-" {
		if input, err = os.Open(inpath); err != nil {
			return nil, err
		}
	}

	if outpath != "-" {
		if output, err = os.Create(outpath); err != nil {
			return nil, err
		}
	}

	return &ioDelegate{input, output, nil, false}, nil
}

func (d *ioDelegate) Read(p []byte) (int, error) {
	return d.infile.Read(p)
}

func (d *ioDelegate) Write(p []byte) (int, error) {
	if d.cache != nil {
		n, err := d.cache.Write(p)
		if err != nil {
			return n, err
		}
	}
	n, err := d.outfile.Write(p)
	return n, err
}

func (d *ioDelegate) TryCache(h hash.Hash, data []byte) (bool, error) {
	dir, err := gtsCacheDir()
	if err != nil {
		return false, nil
	}

	if d.infile == os.Stdin {
		// Write to a temporary file to enable seeking.
		f, err := ioutil.TempFile("", "gts-tmp-*")
		if err != nil {
			return false, nil
		}

		if _, err := io.Copy(f, os.Stdin); err != nil {
			d.Close()
			return false, err
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			d.Close()
			return false, err
		}

		d.infile = f
		d.tmpin = true
	}

	// Compute the root file and data hash sums.
	h.Reset()
	if _, err := io.Copy(h, d.infile); err != nil {
		if _, err := d.infile.Seek(0, io.SeekStart); err != nil {
			return false, err
		}
		return false, nil
	}
	rsum := h.Sum(nil)

	h.Reset()
	h.Write(data)
	dsum := h.Sum(nil)

	if _, err := d.infile.Seek(0, io.SeekStart); err != nil {
		return false, err
	}

	// Open a cache file, or create one if necessary.
	f, err := cache.Open(dir, h, rsum, dsum)
	if err != nil {
		f, err := cache.CreateLevel(dir, h, rsum, dsum, flate.BestSpeed)
		if err != nil && f != nil {
			os.Remove(f.Name())
		}
		d.cache = f
		return false, nil
	}

	defer f.Close()

	if _, err := io.Copy(d.outfile, f); err != nil {
		return false, nil
	}

	if d.outfile != os.Stdout {
		os.Remove(f.Name())
	}

	return true, nil
}

func (d *ioDelegate) Close() error {
	if d.tmpin {
		defer os.Remove(d.infile.Name())
	}

	defer d.infile.Close()
	defer d.outfile.Close()

	if d.cache != nil {
		if err := d.cache.Close(); err != nil {
			os.Remove(d.cache.Name())
		}
	}

	return nil
}
