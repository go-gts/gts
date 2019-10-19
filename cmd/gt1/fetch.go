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
	register("fetch", "download data from various databases via TogoWS", fetchFunc)
}

func fetchFunc(command *flags.Command, args []string) error {
	uniprot := newFetchEntryFunc("uniprot", "raw", "fasta", "gff", "json", "ttl", "xml")
	ena := newFetchEntryFunc("ena", "raw", "fasta", "gff", "json", "xml")
	ddbj := newFetchEntryFunc("ddbj", "raw", "fasta", "gff", "json", "xml")
	refseqn := newFetchEntryFunc("nucleotide", "raw", "fasta", "gb", "gff", "json", "ttl", "xml")

	command.Command("uniprot", "download a UniProt entry via TogoWS", uniprot)
	command.Command("ena", "download a ENA entry via TogoWS", ena)
	command.Command("ddbj", "download a DDBJ entry via TogoWS", ddbj)
	command.Command("refseqn", "download a RefSeq nucleotide entry via TogoWS", refseqn)

	return command.Run(args)
}

func togowsEntryURL(db string, ids []string, field string, format string) string {
	url := fmt.Sprintf("http://togows.org/entry/%s/%s", db, strings.Join(ids, ","))
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

func newFetchEntryFunc(db string, formats ...string) flags.CommandFunc {
	return func(command *flags.Command, args []string) error {
		outfile := command.Outfile("output file")
		id := command.Positional.String("id", "entry ID")
		format := command.Choice('f', "format", "data format to retreive", formats...)
		field := command.String(0, "field", "", "entry field to extract")
		additional := command.Strings(0, "and", "additional entry ids to fetch")

		return command.Run(args, func() error {
			ids := append([]string{*id}, (*additional)...)
			url := togowsEntryURL(db, ids, *field, formats[*format])

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
