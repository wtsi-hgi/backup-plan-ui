package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gocarina/gocsv"
)

const numTestDataRows = 3

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
		if !reflect.DeepEqual(entries[i], originalEntries[i]) {
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

	err := csvSource.updateEntry(newEntry)
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

	if !reflect.DeepEqual(entries[0], newEntry) {
		t.Errorf("First entry does not match the updated entry.\nGot %+v, expected %+v",
			entries[0], newEntry)
	}
}

func TestDeleteEntry(t *testing.T) {
	rowsToTest := []uint16{0, max(0, numTestDataRows-2), numTestDataRows - 1}

	for _, id := range rowsToTest {
		idToDelete := uint16(id)

		t.Run(fmt.Sprintf("Entry %d", idToDelete), func(t *testing.T) {
			entriesBefore, testPath := createTestData(t)

			csvSource := CSVSource{path: testPath}

			err := csvSource.deleteEntry(idToDelete)
			if err != nil {
				t.Fatal(err)
			}

			entriesAfter, err := csvSource.readAll()
			if err != nil {
				t.Fatal(err)
			}

			if len(entriesAfter) != (len(entriesBefore) - 1) {
				t.Errorf("CSV has the wrong number of entries.\nGot %+v, expected %+v",
					entriesAfter[0], entriesBefore[0])
			}

			for _, e := range entriesAfter {
				if e.ID == idToDelete {
					t.Errorf("Deleted entry still present: %+v", e)
				}
			}
		})
	}
}

func createTestData(t *testing.T) ([]*Entry, string) {
	t.Helper()

	baseEntry := Entry{
		ReportingName: "test_project",
		ReportingRoot: "/some/path/to/project/dir",
		Directory:     "/some/path/to/project/dir/input",
		Instruction:   Backup,
		Requestor:     "user",
		Faculty:       "group",
	}

	entries := make([]*Entry, numTestDataRows)

	for i := range numTestDataRows {
		newEntry := baseEntry
		newEntry.ReportingName = fmt.Sprintf("test_project_%d", i)
		newEntry.ID = uint16(i)

		entries[i] = &newEntry
	}

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

func TestWriteEntries(t *testing.T) {
	entries, _ := createTestData(t)

	filePath := filepath.Join(t.TempDir(), "test.csv")
	csvSource := CSVSource{path: filePath}

	err := csvSource.writeEntries(entries)
	if err != nil {
		t.Fatal(err)
	}

	newEntries, err := csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(newEntries) != numTestDataRows {
		t.Fatal("Number of read entries is incorrect")
	}

	if !reflect.DeepEqual(newEntries, entries) {
		t.Errorf("Written entry does not match expected entries.\nGot %+v, expected %+v", newEntries, entries)
	}
}

func TestAddEntry(t *testing.T) {
	entries, filePath := createTestData(t)
	csvSource := CSVSource{path: filePath}

	newEntry := entries[0]
	newEntry.ReportingName = "test_project_new"

	err := csvSource.addEntry(newEntry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err = csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != numTestDataRows+1 {
		t.Fatal("Number of read entries is incorrect")
	}

	if !reflect.DeepEqual(entries[numTestDataRows], newEntry) {
		t.Errorf("New entry does not match expected values.\nGot %+v, expected %+v", entries[numTestDataRows], newEntry)
	}
}

func TestGetNextID(t *testing.T) {
	tests := []struct {
		entries    []*Entry
		expectedID uint16
	}{
		{
			entries:    []*Entry{&Entry{ID: 0}, &Entry{ID: 1}, &Entry{ID: 2}},
			expectedID: 3,
		},
		{
			entries:    []*Entry{&Entry{ID: 0}, &Entry{ID: 2}, &Entry{ID: 3}},
			expectedID: 1,
		},
		{
			entries:    []*Entry{&Entry{ID: 1}, &Entry{ID: 5}, &Entry{ID: 6}},
			expectedID: 0,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Expect %d", test.expectedID), func(t *testing.T) {
			csvSource := CSVSource{}
			id := csvSource.getNextID(test.entries)
			if id != test.expectedID {
				t.Errorf("Expected ID %d, got %d", test.expectedID, id)
			}
		})
	}
}
