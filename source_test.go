package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gocarina/gocsv"
)

func TestReadAll(t *testing.T) {
	originalEntries, testPath := createTestData(t)

	csvSource := CSVSource{path: testPath}

	entries, err := csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != len(originalEntries) {
		t.Fatalf("Number of read entries is incorrect. \n Got %+v, expected %+v",
			len(entries), len(originalEntries))
	}

	for i := range entries {
		if !reflect.DeepEqual(*entries[i], originalEntries[i]) {
			t.Errorf("Entry %d mismatch.\nGot %+v, expected %+v",
				i, entries[i], originalEntries[i])
		}
	}
}

func TestUpdateEntry(t *testing.T) {
	originalEntries, testPath := createTestData(t)

	csvSource := CSVSource{path: testPath}

	newEntry := originalEntries[0]
	newEntry.ReportingName = "test_project_updated"

	err := csvSource.updateEntry(&newEntry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != len(originalEntries) {
		t.Errorf("CSV has the wrong number of entries.\nGot %+v, expected %+v",
			entries[0], originalEntries[0])
	}

	if !reflect.DeepEqual(*entries[0], newEntry) {
		t.Errorf("First entry does not match the updated entry.\nGot %+v, expected %+v",
			entries[0], newEntry)
	}
}

func TestDeleteEntry(t *testing.T) {
	var tests = []struct {
		idToDelete uint16
	}{
		{0},
		{1},
		{2},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("Entry %d", tt.idToDelete)
		t.Run(testname, func(t *testing.T) {
			originalEntries, testPath := createTestData(t)
		
			csvSource := CSVSource{path: testPath}

			err := csvSource.deleteEntry(tt.idToDelete)
			if err != nil {
				t.Fatal(err)
			}

			entries, err := csvSource.readAll()
			if err != nil {
				t.Fatal(err)
			}

			if len(entries) != (len(originalEntries) - 1) {
				t.Errorf("CSV has the wrong number of entries.\nGot %+v, expected %+v",
					entries[0], originalEntries[0])
			}

			for _, e := range entries {
				if e.ID == tt.idToDelete {
					t.Errorf("Deleted entry still present: %+v", e)
				}
			}
		})
	}
}

func createTestData(t *testing.T) ([]Entry, string) {
	t.Helper()

	baseEntry := Entry{
		ReportingName: "test_project",
		ReportingRoot: "/path/to/project",
		Directory:     "/path/to/project/input",
		Instruction:   Backup,
		Requestor:     "user",
		Faculty:       "group",
	}

	var entries []Entry
	for i := range 3 {
		newEntry := baseEntry
		newEntry.ReportingName = fmt.Sprintf("test_project_%d", i)
		newEntry.ID = uint16(i)

		entries = append(entries, newEntry)
	}

	dir := os.TempDir()
	testPath := filepath.Join(dir, "testing")

	out, err := os.Create(testPath)
	if err != nil {
		t.Fatal(err)
	}

	defer out.Close()

	err = gocsv.MarshalFile(&entries, out)
	if err != nil {
		t.Fatal(err)
	}

	return entries, testPath
}
