package converter

import (
	"errors"
	"os"
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

	t.Cleanup(func() {
		err = sq.Close()
		if err != nil {
			t.Log(err)
		}
	})

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

func TestConvertCsvToMySQL(t *testing.T) {
	entries, csvPath := sources.CreateTestCSV(t)

	tableName := "test_convert"

	sq, err := sources.NewMySQLSource(
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_DATABASE"),
		tableName,
	)
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

	err = ConvertCsvToMySQL(
		csvPath,
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_DATABASE"),
		tableName,
	)

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
