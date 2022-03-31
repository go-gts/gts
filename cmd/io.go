package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd/cache"
)

func CacheDir(namespace string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return dir, err
	}
	dir = filepath.Join(dir, namespace)
	return dir, os.MkdirAll(dir, 0755)
}

type MultiError []error

func NewMultiError(errs []error) MultiError {
	for _, err := range errs {
		if err != nil {
			return errs
		}
	}
	return nil
}

func (errs MultiError) Error() string {
	ss := make([]string, len(errs))
	for i, err := range errs {
		ss[i] = fmt.Sprint(err)
	}
	return strings.Join(ss, "\n")
}

type multiWriteCloser []io.WriteCloser

func MultiWriteCloser(wcs ...io.WriteCloser) io.WriteCloser {
	return multiWriteCloser(wcs)
}

func (mwc multiWriteCloser) Write(p []byte) (int, error) {
	errs := make([]error, len(mwc))
	for i, wc := range mwc {
		n, err := wc.Write(p)
		if err != nil {
			errs[i] = err
		}
		if n != len(p) {
			errs[i] = io.ErrShortWrite
		}
	}
	return len(p), NewMultiError(errs)
}

func (mwc multiWriteCloser) Close() error {
	errs := make([]error, len(mwc))
	for i, wc := range mwc {
		errs[i] = wc.Close()
	}
	return NewMultiError(errs)
}

type TempFile struct {
	f *os.File
}

func NewTempFile(prefix string) (*TempFile, error) {
	f, err := ioutil.TempFile("", fmt.Sprintf("%s-*", prefix))
	if err != nil {
		return nil, err
	}
	return &TempFile{f}, nil
}

func (tf *TempFile) Seek(offset int64, whence int) (int64, error) {
	return tf.f.Seek(offset, whence)
}

func (tf *TempFile) Read(p []byte) (int, error) {
	return tf.f.Read(p)
}

func (tf *TempFile) Write(p []byte) (int, error) {
	return tf.f.Write(p)
}

func (tf *TempFile) Close() error {
	if err := tf.f.Close(); err != nil {
		return err
	}
	if err := os.Remove(tf.f.Name()); err != nil {
		return err
	}
	return nil
}

var Base64Encode = base64.RawURLEncoding.WithPadding(base64.NoPadding).EncodeToString

var ErrCacheUsed = errors.New("used cache")

func SeekStart(s io.Seeker) error {
	offset, err := s.Seek(0, io.SeekStart)
	if err == nil && offset != 0 {
		err = errors.New("failed to seek to offset 0")
	}
	return err
}

type IOConfig struct {
	Namespace string
	InPath    string
	OutPath   string
	Payload   cache.Payload
}

func NewIOConfig(namespace, inpath, outpath string) IOConfig {
	return IOConfig{
		Namespace: namespace,
		InPath:    inpath,
		OutPath:   outpath,
	}
}

func (cfg IOConfig) WithPayload(payload cache.Payload) IOConfig {
	return IOConfig{
		Namespace: cfg.Namespace,
		InPath:    cfg.InPath,
		OutPath:   cfg.OutPath,
		Payload:   payload,
	}
}

func HandleIO(cfg IOConfig) (io.ReadCloser, io.WriteCloser, error) {
	r, err := func() (io.ReadSeekCloser, error) {
		// Open the file if STDIN is not expected.
		if cfg.InPath != "-" {
			return os.Open(cfg.InPath)
		}

		// Try to open a new temporary file.
		f, err := NewTempFile(cfg.Namespace)
		if err != nil {
			return os.Stdin, nil
		}

		// Copy STDIN to the temporary file.
		if _, err := io.CopyBuffer(f, os.Stdin, nil); err != nil {
			return nil, err
		}

		// Seek to the start of the temporary file.
		ret, err := f.Seek(0, io.SeekStart)
		if ret != 0 {
			err = errors.New("could not seek to offset 0")
		}

		return f, err
	}()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open input file: %v", err)
	}

	var w io.WriteCloser = os.Stdout
	if cfg.OutPath != "-" {
		if w, err = os.Create(cfg.OutPath); err != nil {
			return nil, nil, fmt.Errorf("failed to open output file: %v", err)
		}
	}

	if cfg.Payload == nil {
		return r, w, nil
	}

	data, err := cache.EncodePayload(cfg.Payload)
	if err != nil {
		gts.Debugf("failed to encode payload: %v", err)
		return r, w, nil
	}

	h := sha256.New224()
	h.Reset()
	h.Write(data)
	dsum := Base64Encode(h.Sum(nil))

	h.Reset()
	if _, err := io.CopyBuffer(h, r, nil); err != nil {
		gts.Debugf("failed to hash root file: %v", err)
		return r, w, SeekStart(r)
	}
	rsum := Base64Encode(h.Sum(nil))

	if err := SeekStart(r); err != nil {
		return nil, nil, err
	}

	dir, err := CacheDir(cfg.Namespace)
	if err != nil {
		gts.Debugf("in cache directory retrievial: %v", err)
		return r, w, nil
	}

	name := filepath.Join(dir, fmt.Sprintf("%s-%s", rsum, dsum))
	if f, err := os.Open(name); err == nil {
		if _, err := io.CopyBuffer(w, f, nil); err != nil {
			gts.Debugf("failed to copy from cache: %v", err)
			// Retry without cache.
			r.Close()
			w.Close()
			return HandleIO(cfg.WithPayload(nil))
		}
		return nil, nil, ErrCacheUsed
	}

	if f, err := os.Create(name); err == nil {
		return r, MultiWriteCloser(f, w), nil
	}

	return r, w, nil
}
