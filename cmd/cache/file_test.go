package cache

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

func makeSum(t *testing.T, h hash.Hash) []byte {
	t.Helper()
	h.Reset()
	if _, err := io.CopyN(h, rand.Reader, int64(h.Size())); err != nil {
		t.Fatal(err)
	}
	return h.Sum(nil)
}

func TestFile(t *testing.T) {
	s := "sumomomomomomomomomnouchimomomosumomomomomonouchi"
	path := t.TempDir()
	h := sha1.New()
	rsum := makeSum(t, h)
	dsum := makeSum(t, h)
	h.Reset()
	if _, err := h.Write(append(rsum, dsum...)); err != nil {
		t.Fatal(err)
	}
	lsum := h.Sum(nil)

	switch f, err := Create(path, h, rsum, dsum); err {
	case nil:
		if _, err := io.WriteString(f, s); err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(path, hex.EncodeToString(lsum))
		if f.Name() != name {
			t.Fatalf("f.Name() = %q, want %q", f.Name(), name)
		}
		if f.ReadOnly() {
			t.Fatal("expected file to be read-write")
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

	default:
		t.Fatal(err)
	}

	switch f, err := Open(path, h, rsum, dsum); err {
	case nil:
		p, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(path, hex.EncodeToString(lsum))
		if f.Name() != name {
			t.Fatalf("f.Name() = %q, want %q", f.Name(), name)
		}
		if !f.ReadOnly() {
			t.Fatal("expected file to be read-only")
		}
		if string(p) != s {
			testutils.Diff(t, string(p), s)
		}
		if _, err := f.Write(nil); err != io.EOF {
			t.Fatal("expected Write on read-only file to return io.EOF")
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

	default:
		t.Fatal(err)
	}
}

func TestFileFail(t *testing.T) {
	h := sha1.New()
	size := h.Size()
	path := t.TempDir()
	rsum := makeSum(t, h)
	dsum := makeSum(t, h)

	t.Run("Open", func(t *testing.T) {
		t.Run("no file", func(t *testing.T) {
			if _, err := Open(path, h, rsum, dsum); err == nil {
				t.Fatal("should not be able to open file")
			}
		})

		t.Run("empty file", func(t *testing.T) {
			h.Reset()
			h.Write(append(rsum, dsum...))
			lsum := h.Sum(nil)
			name := filepath.Join(path, hex.EncodeToString(lsum))
			switch f, err := os.Create(name); err {
			case nil:
				f.Close()

			default:
				t.Fatal(err)
			}

			if _, err := Open(path, h, rsum, dsum); err == nil {
				t.Fatal("should not be able to read header")
			}
		})

		t.Run("root hash mismatch", func(t *testing.T) {
			h.Reset()
			h.Write(append(rsum, dsum...))
			lsum := h.Sum(nil)
			name := filepath.Join(path, hex.EncodeToString(lsum))
			switch f, err := os.Create(name); err {
			case nil:
				p := make([]byte, size*3)
				if _, err := f.Write(p); err != nil {
					t.Fatal(err)
				}
				f.Close()

			default:
				t.Fatal(err)
			}

			if _, err := Open(path, h, rsum, dsum); err == nil {
				t.Fatal("root hash should mismatch")
			}
		})

		t.Run("data hash mismatch", func(t *testing.T) {
			h.Reset()
			h.Write(append(rsum, dsum...))
			lsum := h.Sum(nil)
			name := filepath.Join(path, hex.EncodeToString(lsum))
			switch f, err := os.Create(name); err {
			case nil:
				p := make([]byte, size*3)
				copy(p[:size], rsum)
				if _, err := f.Write(p); err != nil {
					t.Fatal(err)
				}
				f.Close()

			default:
				t.Fatal(err)
			}

			if _, err := Open(path, h, rsum, dsum); err == nil {
				t.Fatal("data hash should mismatch")
			}
		})

		t.Run("body hash mismatch", func(t *testing.T) {
			h.Reset()
			h.Write(append(rsum, dsum...))
			lsum := h.Sum(nil)
			name := filepath.Join(path, hex.EncodeToString(lsum))
			switch f, err := os.Create(name); err {
			case nil:
				p := make([]byte, size*3)
				copy(p[:size], rsum)
				copy(p[size:], dsum)
				if _, err := f.Write(p); err != nil {
					t.Fatal(err)
				}
				f.Close()

			default:
				t.Fatal(err)
			}

			if _, err := Open(path, h, rsum, dsum); err == nil {
				t.Fatal("body hash should mismatch")
			}
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("no file", func(t *testing.T) {
			dir := filepath.Join(path, "dir")
			if _, err := Create(dir, h, rsum, dsum); err == nil {
				t.Fatal("should not be able to create file")
			}
		})

		t.Run("bad level", func(t *testing.T) {
			if _, err := CreateLevel(path, h, rsum, dsum, -3); err == nil {
				t.Fatal("should not be able to use level -3")
			}
		})
	})
}
