package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	flags "gopkg.in/flags.v1"
)

type URL = url.URL

const Host = "togows.org"

func get(url URL, w io.Writer) error {
	res, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		p, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s\n%s", res.Status, string(p))
	}

	_, err = bufio.NewReader(res.Body).WriteTo(w)
	return err
}

func getNewline(url URL, w io.Writer) error {
	if err := get(url, w); err != nil {
		return err
	}
	_, err := w.Write([]byte{'\n'})
	return err
}

func init() {
	flags.Add("entry", "download data through the TogoWS REST service", Entry)
	flags.Add("search", "search data through the TogoWS REST service", Search)
	flags.Add("listdb", "list the available databases for entry or search", ListDB)
}

func Entry(ctx *flags.Context) error {
	pos, opt := flags.Args()

	outfile := pos.Output("output file")
	db := pos.String("database", "the database name (or alias) to access")
	id := pos.String("identifier", "the identifier of the data to download")

	idlist := opt.StringSlice('a', "and", nil, "additional identifier(s)")
	fields := opt.StringSlice('f', "field", nil, "name of field to extract")
	format := opt.String('F', "format", "", "output data format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	ext := path.Ext(outfile.Name())
	if ext != "" && *format == "" {
		*format = ext[1:]
	}

	ids := strings.Join(append([]string{*id}, *idlist...), ",")
	url := URL{
		Scheme: "http",
		Host:   Host,
		Path:   path.Join(append([]string{"entry", *db, ids}, *fields...)...),
	}
	if *format != "" {
		url.Path += "." + *format
	}

	return ctx.Raise(get(url, outfile))
}

func Search(ctx *flags.Context) error {
	pos, opt := flags.Args()

	outfile := pos.Output("output file")
	db := pos.String("database", "the database name (or alias) to access")
	query := pos.String("query", "search query string")

	offset := opt.Int('s', "offset", 1, "pagination offset position")
	limit := opt.Int('n', "limit", 100, "pagination result limit")
	count := opt.Switch('c', "count", "report the number of total search results")
	format := opt.String('F', "format", "", "output data format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	ext := path.Ext(outfile.Name())
	if ext != "" && *format == "" {
		*format = ext[1:]
	}

	url := URL{
		Scheme: "http",
		Host:   Host,
		Path:   path.Join("search", *db, *query),
	}

	if *count {
		url.Path = path.Join(url.Path, "count")
	} else {
		url.Path = path.Join(url.Path, fmt.Sprintf("%d,%d", *offset, *limit))
		if *format != "" {
			url.Path += "." + *format
		}
	}

	return ctx.Raise(getNewline(url, outfile))
}

func ListDB(ctx *flags.Context) error {
	pos, opt := flags.Args()

	tool := pos.String("tool", "TogoWS tool name (entry or search)")
	outfile := pos.Output("output file")
	format := opt.String('F', "format", "", "output data format")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	ext := path.Ext(outfile.Name())
	if ext != "" && *format == "" {
		*format = ext[1:]
	}

	url := URL{Scheme: "http", Host: Host, Path: *tool}
	if *format != "" {
		url.Path += "." + *format
	}

	return ctx.Raise(getNewline(url, outfile))
}

func main() {
	name, desc := "togo", "access various biological databases with TogoWS"
	os.Exit(flags.Run(name, desc, flags.Compile()))
}
