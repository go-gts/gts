package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/go-gts/flags"
	"github.com/go-gts/gts/cmd/cache"
)

func init() {
	cacheSet := flags.CommandSet{}

	cacheSet.Register("list", "list the cache files", cacheListFunc)
	cacheSet.Register("path", "print the cache directory path", cachePathFunc)
	cacheSet.Register("purge", "delete all cache files", cachePurgeFunc)

	flags.Register("cache", "manage gts cache files", cacheSet.Compile())
}

func cacheListFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()
	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	dir, err := gtsCacheDir()
	if err != nil {
		return nil
	}

	total := uint64(0)

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		h := newHash()
		hd, err := cache.ReadHeader(f, h.Size())

		h.Reset()
		h.Write(append(hd.RootSum, hd.DataSum...))
		lsum := h.Sum(nil)
		lhex := hex.EncodeToString(lsum)

		base := filepath.Base(info.Name())
		size := uint64(info.Size())

		if lhex != base {
			return fmt.Errorf("in cache %s: Leaf hash and filename mismatch", path)
		}

		total += size

		fmt.Printf("%s\t%s\n", base, humanize.IBytes(size))

		return nil
	}

	if err := filepath.Walk(dir, walker); err != nil {
		return ctx.Raise(err)
	}

	fmt.Printf("Total\t%s\n", humanize.IBytes(total))

	return nil
}

func cachePathFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()
	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	dir, err := gtsCacheDir()
	if err != nil {
		return nil
	}

	fmt.Println(dir)

	return nil
}

func cachePurgeFunc(ctx *flags.Context) error {
	pos, opt := flags.Flags()
	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	dir, err := gtsCacheDir()
	if err != nil {
		return nil
	}

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path == dir {
				return nil
			}
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			return filepath.SkipDir
		}

		return os.Remove(path)
	}

	if err := filepath.Walk(dir, walker); err != nil {
		return ctx.Raise(err)
	}

	return nil
}
