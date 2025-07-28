package main

import (
	"errors"
	"os"

	"github.com/gocarina/gocsv"
)

type DataSource interface {
	readAll() ([]*Entry, error)
	getEntry(id uint16) (*Entry, error)
	updateEntry(newEntry *Entry) error
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

	err = gocsv.UnmarshalFile(in, &entries)

	return entries, err
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

func (c CSVSource) updateEntry(newEntry *Entry) error {
	entries, err := c.readAll()
	if err != nil {
		return err
	}

	found := false
	for i, entry := range entries {
		if entry.ID == newEntry.ID {
			entries[i] = newEntry
			found = true

			break
		}
	}

	if !found {
		return ErrNoEntry
	}

	out, err := os.Create(c.path)
	if err != nil {
		return err
	}

	defer out.Close()

	return gocsv.MarshalFile(&entries, out)
}
