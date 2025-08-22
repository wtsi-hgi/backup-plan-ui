package converter

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/wtsi-hgi/backup-plan-ui/sources"

	. "github.com/smarty/assertions"
)

func TestConvertCsvToSqlite(t *testing.T) {
	entries, csvPath := sources.CreateTestCSV(t)
	sqlitePath := filepath.Join(t.TempDir(), "test.sqlite")

	sq, err := sources.NewSQLiteSource(sqlitePath)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err = sq.Close()
		if err != nil {
			t.Log(err)
		}
	})

	t.Run("Check entries", func(t *testing.T) {
		err = ConvertCsvToSqlite(csvPath, sqlitePath, true)
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
	})

	testCases := []struct {
		name       string
		dropTable  bool
		numEntries int
	}{
		{"Keep table", false, 2 * len(entries)},
		{"Overwrite table", true, len(entries)},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err = ConvertCsvToSqlite(csvPath, sqlitePath, tt.dropTable)
			if err != nil {
				t.Fatal(err)
			}

			newEntries, err := sq.ReadAll()
			if err != nil {
				t.Fatal(err)
			}

			if ok, e := So(newEntries, ShouldHaveLength, tt.numEntries); !ok {
				t.Error(e)
			}
		})
	}
}

func TestConvertCsvToMySQL(t *testing.T) {
	entries, csvPath := sources.CreateTestCSV(t)

	tableName := "test_convert"

	sq, err := sources.NewMySQLSourceFromEnv(tableName)
	if err != nil {
		if errors.Is(err, sources.ErrMissingArgument) {
			t.Skip("Skipping MySQL test because MySQL host, port, user, pass, or database is not set.")
		}

		t.Fatal(err)
	}

	t.Cleanup(func() {
		err = sq.DropTable()
		if err != nil {
			t.Log(err)
		}

		err = sq.Close()
		if err != nil {
			t.Log(err)
		}
	})

	t.Run("Check entries", func(t *testing.T) {
		err = ConvertCsvToMySQL(csvPath, tableName, true)
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
	})

	testCases := []struct {
		name       string
		dropTable  bool
		numEntries int
	}{
		{"Keep table", false, 2 * len(entries)},
		{"Overwrite table", true, len(entries)},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err = ConvertCsvToMySQL(csvPath, tableName, tt.dropTable)
			if err != nil {
				t.Fatal(err)
			}

			newEntries, err := sq.ReadAll()
			if err != nil {
				t.Fatal(err)
			}

			if ok, e := So(newEntries, ShouldHaveLength, tt.numEntries); !ok {
				t.Error(e)
			}
		})
	}
}
