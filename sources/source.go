package sources

import (
	"errors"
	"os"

	"github.com/gocarina/gocsv"
)

type DataSource interface {
	ReadAll() ([]*Entry, error)
	GetEntry(id uint16) (*Entry, error)
	UpdateEntry(newEntry *Entry) error
	DeleteEntry(id uint16) (*Entry, error)
	AddEntry(entry *Entry) error
}

type CSVSource struct {
	Path string
}

type Instruction string

const (
	Backup     Instruction = "backup"
	NoBackup   Instruction = "nobackup"
	TempBackup Instruction = "tempbackup"
)

type Entry struct {
	ReportingName string      `csv:"reporting_name"`
	ReportingRoot string      `csv:"reporting_root"`
	Directory     string      `csv:"directory"`
	Instruction   Instruction `csv:"instruction"`
	Match         string      `csv:"match"`
	Ignore        string      `csv:"ignore"`
	Requestor     string      `csv:"requestor"`
	Faculty       string      `csv:"faculty"`
	ID            uint16      `csv:"id"`
}

var ErrNoEntry = errors.New("entry does not exist")

func (c CSVSource) ReadAll() ([]*Entry, error) {
	in, err := os.Open(c.Path)
	if err != nil {
		return nil, err
	}

	defer in.Close()

	entries := []*Entry{}

	err = gocsv.UnmarshalFile(in, &entries)

	return entries, err
}

func (c CSVSource) GetEntry(id uint16) (*Entry, error) {
	entries, err := c.ReadAll()
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

func (c CSVSource) UpdateEntry(newEntry *Entry) error {
	entries, err := c.ReadAll()
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
	out, err := os.Create(c.Path)
	if err != nil {
		return err
	}

	defer out.Close()

	return gocsv.MarshalFile(&entries, out)
}

func (c CSVSource) DeleteEntry(id uint16) (*Entry, error) {
	entries, err := c.ReadAll()
	if err != nil {
		return nil, err
	}

	entry, index, err := getMatchingEntryWithID(id, entries)
	if err != nil {
		return nil, err
	}

	entries = append(entries[:index], entries[index+1:]...)

	return entry, c.writeEntries(entries)
}

func (c CSVSource) AddEntry(newEntry *Entry) error {
	entries, err := c.ReadAll()
	if err != nil {
		return err
	}

	newEntry.ID = c.getNextID(entries)

	entries = append(entries, newEntry)

	return c.writeEntries(entries)
}

func (c CSVSource) getNextID(entries []*Entry) uint16 {
	used := make(map[uint16]struct{}, len(entries))
	for _, entry := range entries {
		used[entry.ID] = struct{}{}
	}

	// Find gaps
	for i := range uint16(len(used)) {
		_, found := used[i]
		if !found {
			return i
		}
	}

	return uint16(len(used))
}
