package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type dbName struct {
	Name  string
	Alias string
}

func getEntryDBList() ([]dbName, error) {
	res, err := http.Get("http://togows.org/entry.json")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	decoder := json.NewDecoder(res.Body)

	ret := new([]dbName)
	if err := decoder.Decode(ret); err != nil {
		return nil, err
	}

	return *ret, nil
}

func getEntryFields(db string) ([]string, error) {
	url := fmt.Sprintf("http://togows.org/entry/%s.json?fields", db)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	decoder := json.NewDecoder(res.Body)

	ret := new([]string)
	if err := decoder.Decode(ret); err != nil {
		return nil, err
	}

	return *ret, nil
}

func getEntryFormats(db string) ([]string, error) {
	url := fmt.Sprintf("http://togows.org/entry/%s.json?formats", db)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	decoder := json.NewDecoder(res.Body)

	ret := new([]string)
	if err := decoder.Decode(ret); err != nil {
		return nil, err
	}

	return *ret, nil
}

func getSearchDBList() ([]dbName, error) {
	res, err := http.Get("http://togows.org/search.json")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	decoder := json.NewDecoder(res.Body)

	ret := new([]dbName)
	if err := decoder.Decode(ret); err != nil {
		return nil, err
	}

	return *ret, nil
}
