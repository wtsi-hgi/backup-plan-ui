package converter

import (
	"path/filepath"
	"testing"

	"backup-plan-ui/sources"

	. "github.com/smarty/assertions"
)

func TestConvertCsvToSqlite(t *testing.T) {
	entries, csvPath := sources.CreateTestCSV(t)
	sqlitePath := filepath.Join(t.TempDir(), "test.sqlite")

	err := ConvertCsvToSqlite(csvPath, sqlitePath)
	if err != nil {
		t.Fatal(err)
	}

	sq, err := sources.NewSQLiteSource(sqlitePath)
	if err != nil {
		t.Fatal(err)
	}

	newEntries, err := sq.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		entry.ID += 1
	}

	if ok, e := So(newEntries, ShouldResemble, entries); !ok {
		t.Error(e)
	}
}
