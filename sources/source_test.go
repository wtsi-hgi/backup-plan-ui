package sources

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smarty/assertions"
)

func testDataSourceReadAll(t *testing.T, ds DataSource, originalEntries []*Entry) {
	t.Helper()

	entries, err := ds.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := So(entries, ShouldHaveLength, len(originalEntries)); !ok {
		t.Fatal(err)
	}

	for i := range entries {
		if ok, err := So(entries[i], ShouldResemble, originalEntries[i]); !ok {
			t.Error(err)
		}
	}
}

func testDataSourceGetEntry(t *testing.T, ds DataSource, originalEntries []*Entry) {
	t.Helper()

	for _, originalEntry := range originalEntries {
		entry, err := ds.GetEntry(originalEntry.ID)
		if err != nil {
			t.Fatal(err)
		}

		if ok, err := So(entry, ShouldResemble, originalEntry); !ok {
			t.Error(err)
		}
	}
}

func TestUpdateEntry(t *testing.T) {
	originalEntries, testPath := CreateTestCSV(t)

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
			entriesBefore, testPath := CreateTestCSV(t)

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

func TestAddEntry(t *testing.T) {
	entries, filePath := CreateTestCSV(t)
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
