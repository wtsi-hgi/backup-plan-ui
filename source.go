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
	deleteEntry(id uint16) error
	addEntry(entry *Entry) error
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

	entry, _, err := getMatchingEntryWithID(id, entries)

	return entry, err
}

func getMatchingEntryWithID(id uint16, entries []*Entry) (*Entry, int, error) {
	for i, entry := range entries {
		if entry.ID == uint16(id) {
			return entry, i, nil
		}
	}

	return nil, 0, ErrNoEntry
}

func (c CSVSource) updateEntry(newEntry *Entry) error {
	entries, err := c.readAll()
	if err != nil {
		return err
	}

	_, index, err := getMatchingEntryWithID(newEntry.ID, entries)
	if err != nil {
		return err
	}

	entries[index] = newEntry

	return c.writeEntries(entries)
}

func (c CSVSource) writeEntries(entries []*Entry) error {
	out, err := os.Create(c.path)
	if err != nil {
		return err
	}

	defer out.Close()

	return gocsv.MarshalFile(&entries, out)
}

func (c CSVSource) deleteEntry(id uint16) error {
	entries, err := c.readAll()
	if err != nil {
		return err
	}

	_, index, err := getMatchingEntryWithID(id, entries)
	if err != nil {
		return err
	}

	entries = append(entries[:index], entries[index+1:]...)

	return c.writeEntries(entries)
}

func (c CSVSource) addEntry(newEntry *Entry) error {
	entries, err := c.readAll()
	if err != nil {
		return err
	}

	newEntry.ID = uint16(len(entries))

	entries = append(entries, newEntry)

	return c.writeAll(entries)
}

func (c CSVSource) writeAll(entries []*Entry) error {
	file, err := os.Create(c.path)
	if err != nil {
		return err
	}
	defer file.Close()

	return gocsv.MarshalFile(entries, file)
}
