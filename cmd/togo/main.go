package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-gts/gts/flags"
)

// Host is the togows host url.
const Host = "togows.org"

// LastByte stores the last byte of a io.Writer Write.
type LastByte byte

// Write satisfies the io.Writer interface.
func (c *LastByte) Write(p []byte) (int, error) {
	if len(p) > 0 {
		*c = LastByte(p[len(p)-1])
	}
	return len(p), nil
}

func cacheFile(addr url.URL, open func(string) (*os.File, error)) (*os.File, bool) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, false
	}
	path := filepath.Join(dir, addr.Host, filepath.FromSlash(addr.Path))
	if os.MkdirAll(filepath.Dir(path), 0755) != nil {
		return nil, false
	}
	f, err := open(path)
	if err != nil {
		return nil, false
	}
	if err != nil {
		return nil, false
	}
	return f, true
}

func download(addr url.URL, w io.Writer) error {
	res, err := http.Get(addr.String())
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		p, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s\n%s", res.Status, string(p))
	}
	if f, ok := cacheFile(addr, os.Create); ok {
		defer f.Close()
		w = io.MultiWriter(w, f)
	}
	lb := new(LastByte)
	w = io.MultiWriter(w, lb)
	_, err = io.Copy(w, res.Body)
	if err != nil {
		return err
	}
	if *lb != '\n' {
		_, err = io.WriteString(w, "\n")
	}
	return err
}

func call(addr url.URL, w io.Writer, cache bool) error {
	if cache {
		if f, ok := cacheFile(addr, os.Open); ok {
			defer f.Close()
			_, err := io.Copy(w, f)
			return err
		}
	}
	return download(addr, w)
}

func init() {
	flags.Register("entry", "download database entries through the TogoWS REST service", entry)
	flags.Register("search", "search for database entries through the TogoWS REST service", search)
	flags.Register("listdb", "list databases available for access with entry or search", listdb)
}

func entry(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	db := pos.String("database", "database name (or alias) to access")
	id := pos.String("identifier", "identifier of entry to download")

	outpath := opt.String('o', "output", "-", "file to output (specifying `-` will force standard output)")
	idlist := opt.StringSlice('a', "and", nil, "additional identifier(s) to download")
	fields := opt.StringSlice('f', "field", nil, "name of field(s) to extract")
	format := opt.String('F', "format", "", "output data format")
	nocache := opt.Switch(0, "no-cache", "ignore local cache files")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	outfile := os.Stdout
	if *outpath != "-" {
		f, err := os.Create(*outpath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *outpath, err))
		}
		outfile = f
	}

	ext := path.Ext(outfile.Name())
	if ext != "" && *format == "" {
		*format = ext[1:]
	}

	*idlist = append([]string{*id}, *idlist...)
	ids := strings.Join(*idlist, ",")
	parts := append([]string{"entry", *db, ids}, *fields...)
	pth := path.Join(parts...)
	if *format != "" {
		pth = fmt.Sprintf("%s.%s", pth, *format)
	}

	addr := url.URL{
		Scheme: "http",
		Host:   Host,
		Path:   pth,
	}

	return ctx.Raise(call(addr, outfile, !*nocache))
}

func search(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	db := pos.String("database", "database name (or alias) to access")
	query := pos.String("query", "search query string")

	outpath := opt.String('o', "output", "-", "file to output (specifying `-` will force standard output)")
	offset := opt.Int('s', "offset", 1, "pagination offset position")
	limit := opt.Int('n', "limit", 100, "pagination result limit")
	count := opt.Switch('c', "count", "report the number of total search results")
	format := opt.String('F', "format", "", "output data format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	outfile := os.Stdout
	if *outpath != "-" {
		f, err := os.Create(*outpath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *outpath, err))
		}
		outfile = f
	}

	ext := path.Ext(outfile.Name())
	if ext != "" && *format == "" {
		*format = ext[1:]
	}

	pth := path.Join("search", *db, *query)
	if *count {
		pth = path.Join(pth, "count")
	} else {
		pth = path.Join(pth, fmt.Sprintf("%d,%d", *offset, *limit))
		if *format != "" {
			pth = fmt.Sprintf("%s.%s", pth, *format)
		}
	}

	addr := url.URL{
		Scheme: "http",
		Host:   Host,
		Path:   pth,
	}

	return ctx.Raise(download(addr, outfile))
}

func listdb(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	tool := pos.String("tool", "TogoWS tool name (entry or search")

	outpath := opt.String('o', "output", "-", "file to output (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output data format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	outfile := os.Stdout
	if *outpath != "-" {
		f, err := os.Create(*outpath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *outpath, err))
		}
		outfile = f
	}

	switch *tool {
	case "entry", "search":
	default:
		ctx.Raise(fmt.Errorf("unknown tool name: %s", *tool))
	}

	ext := path.Ext(outfile.Name())
	if ext != "" && *format == "" {
		*format = ext[1:]
	}

	pth := *tool
	if *format != "" {
		pth = fmt.Sprintf("%s.%s", pth, *format)
	}

	addr := url.URL{
		Scheme: "http",
		Host:   Host,
		Path:   pth,
	}

	return ctx.Raise(download(addr, outfile))
}

func main() {
	name, desc := "togo", "access various biological databases with TogoWS"
	version := flags.Version{
		Major: 1,
		Minor: 0,
		Patch: 1,
	}
	os.Exit(flags.Run(name, desc, version, flags.Compile()))
}
