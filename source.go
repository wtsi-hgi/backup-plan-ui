package main

import (
	"errors"
	"os"

	"github.com/gocarina/gocsv"
)

type DataSource interface {
	readAll() ([]*Entry, error)
	getEntry(id uint16) (*Entry, error)
}

type CSVSource struct {
	path string
}

var ErrNoEntry = errors.New("Entry does not exist")

func (c CSVSource) readAll() ([]*Entry, error) {
	in, err := os.Open(c.path)
	if err != nil {
		return nil, err
	}

	defer in.Close()

	entries := []*Entry{}

	if err := gocsv.UnmarshalFile(in, &entries); err != nil {
		return nil, err
	}

	for i, entry := range entries {
		entry.ID = uint16(i)
	}

	return entries, nil
}

func (c CSVSource) getEntry(id uint16) (*Entry, error) {
	entries, err := c.readAll()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.ID == uint16(id) {
			return entry, nil
		}
	}

	return nil, ErrNoEntry
}
