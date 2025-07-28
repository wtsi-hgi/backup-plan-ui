package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gocarina/gocsv"
)

const numTestDataRows = 3

var firstEntry = Entry{
	ReportingName: "test_project",
	ReportingRoot: "/path/to/project",
	Directory:     "/path/to/project/input",
	Instruction:   Backup,
	Requestor:     "user",
	Faculty:       "group",
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return destinationFile.Sync()
}

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

func createTestData(t *testing.T) ([]Entry, string) {
	t.Helper()

	baseEntry := firstEntry
	var entries []Entry
	for i := range numTestDataRows {
		newEntry := baseEntry
		newEntry.ReportingName = fmt.Sprintf("test_project_%d", i)
		newEntry.ID = uint16(i)

		entries = append(entries, newEntry)
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

func TestWriteAll(t *testing.T) {
	entries := []*Entry{&firstEntry}

	filePath := filepath.Join(t.TempDir(), "test.csv")
	defer os.Remove(filePath)

	csvSource := CSVSource{path: filePath}

	err := csvSource.writeAll(entries)
	if err != nil {
		t.Fatal(err)
	}

	newEntries, err := csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(newEntries) != 1 {
		t.Fatal("Number of read entries is incorrect")
	}

	if !reflect.DeepEqual(*newEntries[0], firstEntry) {
		t.Errorf("First entry does not match expected values.\nGot %+v, expected %+v", newEntries[0], firstEntry)
	}
}

func TestAddEntry(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "test.csv")
	defer os.Remove(filePath)

	err := copyFile("data/plan.csv", filePath)
	if err != nil {
		t.Fatal(err)
	}

	csvSource := CSVSource{path: filePath}

	err = csvSource.addEntry(&firstEntry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Fatal("Number of read entries is incorrect")
	}

	if !reflect.DeepEqual(*entries[2], firstEntry) {
		t.Errorf("First entry does not match expected values.\nGot %+v, expected %+v", entries[0], firstEntry)
	}
}
