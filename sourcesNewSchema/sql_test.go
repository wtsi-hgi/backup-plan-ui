package sourcesNewSchema

import (
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	. "github.com/smarty/assertions"
)

var sqlTestCases = []struct {
	name  string
	setup func(t *testing.T) SQLSourceInterface
}{
	{"SQLite", setupSQLiteSourceForTest},
	{"MySQL", setupMySQLSourceForTest},
}

func TestNewSQLiteSource(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "test.db")

	sq, err := NewSQLiteSource(dbFile)
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

func TestNewMySQLSource(t *testing.T) {
	usersTableName := "test_create_users"
	directoriesTableName := "test_create_directories"
	rulesTableName := "test_create_rules"

	sq, err := NewMySQLSourceFromEnv(usersTableName, directoriesTableName, rulesTableName)
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

func setupMySQLSourceForTest(t *testing.T) SQLSourceInterface {
	t.Helper()

	suffix := rand.Int()

	userTableName := fmt.Sprintf("test_user_%d", suffix)
	directoryTableName := fmt.Sprintf("test_directory_%d", suffix)
	ruleTableName := fmt.Sprintf("test_rule_%d", suffix)

	sq, err := NewMySQLSourceFromEnv(userTableName, directoryTableName, ruleTableName)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(callAndLogCleanup(t, sq.Close))
	t.Cleanup(callAndLogCleanup(t, sq.DropTables))

	return sq
}

func setupSQLiteSourceForTest(t *testing.T) SQLSourceInterface {
	t.Helper()

	dbFile := filepath.Join(t.TempDir(), "test.db")

	sq, err := NewSQLiteSource(dbFile)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(callAndLogCleanup(t, sq.Close))

	return sq
}

func TestSQLSourceInterface_AddUser(t *testing.T) {
	for _, tc := range sqlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			sq := tc.setup(t)

			user := User{0, "test_user", "test_faculty", "test_programme"}

			id, err := sq.AddUser(user)
			if err != nil {
				t.Fatal(err)
			}

			if ok, err := So(id, ShouldBeGreaterThan, 0); !ok {
				t.Error(err)
			}
		})
	}
}

func TestSQLSourceInterface_GetUser(t *testing.T) {
	for _, tc := range sqlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			sq := tc.setup(t)

			user := User{0, "test_user", "test_faculty", "test_programme"}

			id, err := sq.AddUser(user)
			if err != nil {
				t.Fatal(err)
			}

			result, err := sq.GetUser(id)
			if err != nil {
				t.Fatal(err)
			}

			if ok, err := So(result, ShouldNotBeNil); !ok {
				t.Fatal(err)
			}

			user.ID = id

			if ok, err := So(result, ShouldResemble, &user); !ok {
				t.Error(err)
			}

		})
	}
}
