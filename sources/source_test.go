package sources

import (
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

func testDataSourceUpdateEntry(t *testing.T, ds DataSource, originalEntries []*Entry) {
	t.Helper()

	newEntry := originalEntries[0]
	newEntry.ReportingName = "test_project_updated"

	err := ds.UpdateEntry(newEntry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := ds.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := So(entries, ShouldHaveLength, len(originalEntries)); !ok {
		t.Fatal(err)
	}

	if ok, err := So(entries[0], ShouldResemble, newEntry); !ok {
		t.Error(err)
	}
}

func testDataSourceDeleteEntry(t *testing.T, ds DataSource, originalEntry *Entry, idToDelete uint16) {
	entry, err := ds.DeleteEntry(idToDelete)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := So(entry, ShouldResemble, originalEntry); !ok {
		t.Error(err)
	}

	entriesAfter, err := ds.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := So(entriesAfter, ShouldHaveLength, NumTestDataRows-1); !ok {
		t.Error(err)
	}

	for _, e := range entriesAfter {
		if e.ID == idToDelete {
			t.Errorf("Deleted entry still present: %+v", e)
		}
	}
}

func testDataSourceAddEntry(t *testing.T, ds DataSource, originalEntries []*Entry) {
	newEntry := originalEntries[0]
	newEntry.ReportingName = "test_project_new"

	err := ds.AddEntry(newEntry)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := ds.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := So(entries, ShouldHaveLength, len(originalEntries)+1); !ok {
		t.Fatal(err)
	}

	newEntry.ID = originalEntries[len(originalEntries)-1].ID + 1

	if ok, err := So(entries[NumTestDataRows], ShouldResemble, newEntry); !ok {
		t.Error(err)
	}
}

func TestConvert(t *testing.T) {
	csvPath := "../data/plan.csv"
	sqlitePath := "../data/plan.sqlite"

	err := convertCsvToSqlite(csvPath, sqlitePath)
	if err != nil {
		t.Fatal(err)
	}
}
