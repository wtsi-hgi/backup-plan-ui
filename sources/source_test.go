package sources

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gocarina/gocsv"
	. "github.com/smarty/assertions"
)

func TestReadAll(t *testing.T) {
	originalEntries, testPath := createTestData(t)

	csvSource := CSVSource{Path: testPath}

	entries, err := csvSource.ReadAll()
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

	csvSource := CSVSource{Path: testPath}

	newEntry := originalEntries[0]
	newEntry.ReportingName = "test_project_updated"

	err := csvSource.UpdateEntry(newEntry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := csvSource.ReadAll()
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
	rowsToTest := []uint16{0, max(0, NumTestDataRows-2), NumTestDataRows - 1}

	for _, idToDelete := range rowsToTest {
		t.Run(fmt.Sprintf("Entry %d", idToDelete), func(t *testing.T) {
			entriesBefore, testPath := createTestData(t)

			csvSource := CSVSource{Path: testPath}

			entry, err := csvSource.DeleteEntry(idToDelete)
			if err != nil {
				t.Fatal(err)
			}

			if ok, err := So(entry, ShouldResemble, entriesBefore[idToDelete]); !ok {
				t.Error(err)
			}

			entriesAfter, err := csvSource.ReadAll()
			if err != nil {
				t.Fatal(err)
			}

			if ok, err := So(entriesAfter, ShouldHaveLength, len(entriesBefore)-1); !ok {
				t.Error(err)
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

	entries := make([]*Entry, NumTestDataRows)

	for i := range NumTestDataRows {
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
	csvSource := CSVSource{Path: filePath}

	err := csvSource.writeEntries(entries)
	if err != nil {
		t.Fatal(err)
	}

	newEntries, err := csvSource.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(newEntries) != NumTestDataRows {
		t.Fatal("Number of read entries is incorrect")
	}

	if !reflect.DeepEqual(newEntries, entries) {
		t.Errorf("Written entry does not match expected entries.\nGot %+v, expected %+v", newEntries, entries)
	}
}

func TestAddEntry(t *testing.T) {
	entries, filePath := createTestData(t)
	csvSource := CSVSource{Path: filePath}

	newEntry := entries[0]
	newEntry.ReportingName = "test_project_new"

	err := csvSource.AddEntry(newEntry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err = csvSource.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != NumTestDataRows+1 {
		t.Fatal("Number of read entries is incorrect")
	}

	if !reflect.DeepEqual(entries[NumTestDataRows], newEntry) {
		t.Errorf("New entry does not match expected values.\nGot %+v, expected %+v", entries[NumTestDataRows], newEntry)
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
