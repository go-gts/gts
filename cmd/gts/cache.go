package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/go-gts/flags"
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
		if _, err := readCacheHeader(f, h); err != nil {
			return err
		}

		base := filepath.Base(info.Name())
		size := uint64(info.Size())

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
