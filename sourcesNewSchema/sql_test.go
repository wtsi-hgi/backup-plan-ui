package sourcesNewSchema

import (
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
		So(tableNames, ShouldContain, tableName)
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
	directoriesTableName := "test_create_directories"
	rulesTableName := "test_create_rules"

	sq, err := NewMySQLSourceFromEnv(directoriesTableName, rulesTableName)
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

	for _, tableName := range []string{directoriesTableName, rulesTableName} {
		So(tableNames, ShouldContain, tableName)
	}
}

func setupMySQLSourceForTest(t *testing.T) SQLSourceInterface {
	t.Helper()

	suffix := rand.Int()

	directoryTableName := fmt.Sprintf("test_directory_%d", suffix)
	ruleTableName := fmt.Sprintf("test_rule_%d", suffix)

	sq, err := NewMySQLSourceFromEnv(directoryTableName, ruleTableName)
	if err != nil {
		if errors.Is(err, ErrMissingArgument) {
			t.Skip("Skipping MySQL test because MySQL host, port, user, pass, or database is not set.")
		}

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

func TestSQLSourceInterface_AddDirectory(t *testing.T) {
	for _, tc := range sqlTestCases {
		Convey(fmt.Sprintf("Given a %s connection", tc.name), t, func() {
			sq := tc.setup(t)

			Convey("You can add a directory", func() {
				directory := Directory{
					Path:      "/path",
					Faculty:   "test",
					Programme: "test",
				}

				id, err := sq.AddDirectory(directory)
				So(err, ShouldBeNil)
				So(id, ShouldBeGreaterThan, 0)

				Convey("You cannot add a directory with the same path", func() {
					directory2 := directory

					directory.Faculty = "test2"
					directory2.Programme = "test2"

					_, err = sq.AddDirectory(directory2)
					So(err, ShouldEqual, ErrDirectoryDuplicate)
				})
			})
		})
	}
}

func TestSQLSourceInterface_GetDirectory(t *testing.T) {
	for _, tc := range sqlTestCases {
		Convey(fmt.Sprintf("Given a %s connection", tc.name), t, func() {
			sq := tc.setup(t)

			Convey("You cannot get non-existent directory", func() {
				_, err := sq.GetDirectory(1)
				So(err, ShouldEqual, ErrNoDirectory)
			})

			Convey("You can get existing directory", func() {
				directory := Directory{
					Path:      "/path",
					Faculty:   "test",
					Programme: "test",
				}

				id, err := sq.AddDirectory(directory)
				So(err, ShouldBeNil)

				result, err := sq.GetDirectory(id)
				So(err, ShouldBeNil)

				directory.ID = id
				So(result, ShouldResemble, &directory)
			})
		})
	}
}
