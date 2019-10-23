package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ktnyt/gt1/flags"
)

func init() {
	register("search", "search databases", searchFunc)
}

func searchFunc(command *flags.Command, args []string) error {
	dblist, err := getSearchDBList()
	if err != nil {
		return nil
	}

	for _, db := range dblist {
		name, alias := db.Name, db.Alias
		f := newSearchFunc(name)
		if alias == "" {
			command.Command(name, fmt.Sprintf("search %s", name), f)
		} else {
			command.Command(alias, fmt.Sprintf("search %s", alias), f)
		}
	}

	return command.Run(args)
}

func togowsSearchURL(db, query string, offset, limit int, format string) string {
	url := fmt.Sprintf(
		"http://togows.org/search/%s/%s/%d,%d",
		db, url.QueryEscape(query), offset+1, limit,
	)
	if format == "raw" {
		format = ""
	}
	if format != "" {
		url += "." + format
	}
	return url
}

func newSearchFunc(db string) flags.CommandFunc {
	return func(command *flags.Command, args []string) error {
		formats, err := getEntryFormats(db)
		if err != nil {
			return err
		}
		formats = append([]string{"raw"}, formats...)

		outfile := command.Outfile("output file")
		query := command.Positional.String("query", "search query string")
		format := command.Choice('f', "format", "data format to retreive", formats...)
		offset := command.Int('o', "offset", 0, "search result offset")
		limit := command.Int('l', "limit", 50, "search result limit")

		return command.Run(args, func() error {
			url := togowsSearchURL(db, *query, *offset, *limit, formats[*format])

			res, err := http.Get(url)
			if err != nil {
				return err
			}

			defer res.Body.Close()

			if res.StatusCode != 200 {
				return errors.New(res.Status)
			}

			if _, err := bufio.NewReader(res.Body).WriteTo(outfile); err != nil {
				return err
			}

			return nil
		})
	}
}
