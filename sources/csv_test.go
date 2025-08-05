package sources

import (
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/smarty/assertions"
)

func TestCSVSource_ReadAll(t *testing.T) {
	originalEntries, testPath := CreateTestCSV(t)

	csvSource := CSVSource{Path: testPath}

	testDataSourceReadAll(t, csvSource, originalEntries)
}

func TestCSVSource_GetEntry(t *testing.T) {
	originalEntries, testPath := CreateTestCSV(t)

	csvSource := CSVSource{Path: testPath}

	testDataSourceGetEntry(t, csvSource, originalEntries)
}

func TestCSVSource_WriteEntries(t *testing.T) {
	entries, _ := CreateTestCSV(t)

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

	if ok, err := So(newEntries, ShouldHaveLength, NumTestDataRows); !ok {
		t.Fatal(err)
	}

	if ok, err := So(newEntries, ShouldResemble, entries); !ok {
		t.Error(err)
	}
}

func TestGetNextID(t *testing.T) {
	tests := []struct {
		entries    []*Entry
		expectedID uint16
	}{
		{
			entries:    []*Entry{{ID: 0}, {ID: 1}, {ID: 2}},
			expectedID: 3,
		},
		{
			entries:    []*Entry{{ID: 0}, {ID: 2}, {ID: 3}},
			expectedID: 1,
		},
		{
			entries:    []*Entry{{ID: 1}, {ID: 5}, {ID: 6}},
			expectedID: 0,
		},
	}

	csvSource := CSVSource{}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Expect %d", test.expectedID), func(t *testing.T) {
			id := csvSource.getNextID(test.entries)
			if ok, err := So(id, ShouldEqual, test.expectedID); !ok {
				t.Error(err)
			}
		})
	}
}
