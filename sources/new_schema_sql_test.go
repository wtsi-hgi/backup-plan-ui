package sources

import (
	"errors"
	"path/filepath"
	"testing"

	. "github.com/smarty/assertions"
)

func TestNewSchemaNewSQLiteSource(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "test.db")

	sq, err := NewSchemaNewSQLiteSource(dbFile)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(callAndLogCleanup(t, sq.Close))

	tableNames, err := sq.ShowTables()
	if err != nil {
		t.Fatal(err)
	}

	for _, tableName := range DefaultTables {
		if ok, err := So(tableNames, ShouldContain, tableName); !ok {
			t.Error(err)
		}
	}
}

func callAndLogCleanup(t *testing.T, f func() error) func() {
	t.Helper()

	return func() {
		err := f()
		if err != nil {
			t.Log(err)
		}
	}
}

func TestNewSchemaNewMySQLSource(t *testing.T) {
	usersTableName := "test_create_users"
	directoriesTableName := "test_create_directories"
	rulesTableName := "test_create_rules"

	sq, err := NewSchemaNewMySQLSourceFromEnv(usersTableName, directoriesTableName, rulesTableName)
	if err != nil {
		if errors.Is(err, ErrMissingArgument) {
			t.Skip("Skipping MySQL test because MySQL host, port, user, pass, or database is not set.")
		}

		t.Fatal(err)
	}

	t.Cleanup(callAndLogCleanup(t, sq.Close))
	t.Cleanup(callAndLogCleanup(t, sq.DropTables))

	tableNames, err := sq.ShowTables()
	if err != nil {
		t.Fatal(err)
	}

	for _, tableName := range []string{usersTableName, directoriesTableName, rulesTableName} {
		if ok, err := So(tableNames, ShouldContain, tableName); !ok {
			t.Error(err)
		}
	}
}
