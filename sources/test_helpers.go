//go:build test

package sources

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/gocarina/gocsv"
)

const NumTestDataRows = 3

func CreateTestEntries(t *testing.T) []*Entry {
	t.Helper()

	baseEntry := Entry{
		ReportingName: "test_project",
		ReportingRoot: "/some/path/to/project/dir",
		Directory:     "/some/path/to/project/dir/input",
		Instruction:   Backup,
		Requestor:     "user",
		Faculty:       "group",
	}

	entries := make([]*Entry, NumTestDataRows)

	for i := range NumTestDataRows {
		newEntry := baseEntry
		newEntry.ReportingName = fmt.Sprintf("test_project_%d", i)
		newEntry.ID = uint16(i)

		entries[i] = &newEntry
	}

	return entries
}

func CreateTestCSV(t *testing.T) ([]*Entry, string) {
	t.Helper()

	entries := CreateTestEntries(t)

	file, err := os.CreateTemp(t.TempDir(), "*.csv")
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	err = gocsv.MarshalFile(&entries, file)
	if err != nil {
		t.Fatal(err)
	}

	return entries, file.Name()
}

func convertCsvToSqlite(csvPath, sqlitePath string) error {
	csv := CSVSource{Path: csvPath}
	entries, err := csv.ReadAll()
	if err != nil {
		return err
	}

	for i, e := range entries {
		if e.Instruction != Backup && e.Instruction != NoBackup && e.Instruction != TempBackup {
			fmt.Printf("Entry %d: %+v\n", i, e)
			return errors.New("wrong entry")
		}
	}

	sq, err := NewSQLiteSource(sqlitePath)
	if err != nil {
		return err
	}

	defer sq.Close()

	err = sq.CreateTable()
	if err != nil {
		return err
	}

	return sq.writeEntries(entries)
}
