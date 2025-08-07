//go:build test

package sources

import (
	"fmt"
	"os"
	"testing"

	"github.com/gocarina/gocsv"
)

const NumTestDataRows = 3

func createTestEntries(t *testing.T) []*Entry {
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

	entries := createTestEntries(t)

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
