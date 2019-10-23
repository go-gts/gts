package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ktnyt/gt1/flags"
)

func init() {
	register("entry", "download data from various databases", entryFunc)
}

func entryFunc(command *flags.Command, args []string) error {
	dblist, err := getEntryDBList()
	if err != nil {
		return err
	}

	for _, db := range dblist {
		name, alias := db.Name, db.Alias

		fields, err := getEntryFields(name)
		if err != nil {
			return err
		}

		f := newEntryFunc(name, fields)
		usage := fmt.Sprintf("download an entry from %s", alias)
		command.Command(alias, usage, f)
	}

	return command.Run(args)
}

func togowsEntryURL(db string, ids []string, field string, format string) string {
	url := fmt.Sprintf("http://togows.org/entry/%s/%s", db, strings.Join(ids, ","))
	if field == "none" {
		field = ""
	}
	if field != "" {
		url += "/" + field
	}
	if format == "raw" {
		format = ""
	}
	if format != "" {
		url += "." + format
	}
	return url
}

func newEntryFunc(db string, fields []string) flags.CommandFunc {
	fields = append([]string{"none"}, fields...)

	return func(command *flags.Command, args []string) error {
		formats, err := getEntryFormats(db)
		if err != nil {
			return err
		}
		formats = append([]string{"raw"}, formats...)

		outfile := command.Outfile("output file")
		id := command.Positional.String("id", "entry ID")
		format := command.Choice('f', "format", "data format to retreive", formats...)
		field := command.Choice('e', "field", "entry field to extract", fields...)
		additional := command.Strings(0, "and", "additional entry ids to download")

		return command.Run(args, func() error {
			ids := append([]string{*id}, (*additional)...)
			url := togowsEntryURL(db, ids, fields[*field], formats[*format])

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
