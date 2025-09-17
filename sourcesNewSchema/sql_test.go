package sourcesNewSchema

import (
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

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
	Convey("Given a database file", t, func() {
		dbFile := filepath.Join(t.TempDir(), "test.db")

		Convey("You can create a SQLite source", func() {
			sq, err := NewSQLiteSource(dbFile)
			So(err, ShouldBeNil)

			t.Cleanup(callAndLogCleanup(t, sq.Close))

			tableNames, err := sq.ShowTables()
			So(err, ShouldBeNil)

			for _, tableName := range DefaultTables {
				So(tableNames, ShouldContain, tableName)
			}
		})
	})
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
	Convey("You can create a MySQL source", t, func() {
		directoriesTableName := "test_create_directories"
		rulesTableName := "test_create_rules"

		sq, err := NewMySQLSourceFromEnv(directoriesTableName, rulesTableName)
		if err != nil && errors.Is(err, ErrMissingArgument) {
			t.Skip("Skipping MySQL test because MySQL host, port, user, pass, or database is not set.")
		}

		So(err, ShouldBeNil)

		t.Cleanup(callAndLogCleanup(t, sq.Close))
		t.Cleanup(callAndLogCleanup(t, sq.DropTables))

		tableNames, err := sq.ShowTables()
		So(err, ShouldBeNil)

		for _, tableName := range []string{directoriesTableName, rulesTableName} {
			So(tableNames, ShouldContain, tableName)
		}
	})
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

func TestSQLSourceInterface(t *testing.T) {
	for _, tc := range sqlTestCases {
		Convey(fmt.Sprintf("Given a %s connection", tc.name), t, func() {
			sq := tc.setup(t)

			directory := Directory{
				Path:      "/path",
				Faculty:   "test",
				Programme: "test",
			}

			Convey("You cannot get a non-existent directory", func() {
				_, err := sq.GetDirectory(1)
				So(err, ShouldEqual, ErrNoDirectory)
			})

			Convey("You cannot claim a non-existent directory", func() {
				err := sq.ClaimDirectory(1, "testUser")
				So(err, ShouldEqual, ErrNoDirectory)
			})

			Convey("You can add a directory", func() {
				id, err := sq.AddDirectory(directory)
				So(err, ShouldBeNil)
				So(id, ShouldBeGreaterThan, 0)

				Convey("You can get a directory", func() {
					result, err := sq.GetDirectory(id)
					So(err, ShouldBeNil)

					directory.ID = id
					So(result, ShouldResemble, &directory)
				})

				Convey("You cannot add a directory with the same path", func() {
					directory2 := directory

					directory.Faculty = "test2"
					directory2.Programme = "test2"

					_, err = sq.AddDirectory(directory2)
					So(err, ShouldEqual, ErrDirectoryDuplicate)
				})

				Convey("You can claim a directory", func() {
					err = sq.ClaimDirectory(id, "testUser")
					So(err, ShouldBeNil)

					result, err := sq.GetDirectory(id)
					So(err, ShouldBeNil)
					So(result.ClaimedBy, ShouldEqual, "testUser")

					Convey("You can reclaim a directory", func() {
						err = sq.ClaimDirectory(id, "testUser2")
						So(err, ShouldBeNil)

						result, err := sq.GetDirectory(id)
						So(err, ShouldBeNil)
						So(result.ClaimedBy, ShouldEqual, "testUser2")
					})
				})

				Convey("And a valid rule", func() {
					rule := Rule{
						BackupType: "backup",
					}

					rule.SetDefaults()

					Convey("You can set a rule for a directory", func() {
						err = sq.SetRule(id, rule)
						So(err, ShouldBeNil)

						rule.ID = 1
						rule.ReviewAt = rule.ReviewAt.Truncate(24 * time.Hour)
						rule.DeleteAt = rule.DeleteAt.Truncate(24 * time.Hour)

						result, err := sq.GetDirectory(id)
						So(err, ShouldBeNil)
						So(result.Rules, ShouldHaveLength, 1)
						So(result.Rules[0], ShouldResemble, &rule)
					})
				})

			})

			Convey("You can add a directory with a claimed user", func() {
				directory.ClaimedBy = "testUser"

				id, err := sq.AddDirectory(directory)
				So(err, ShouldBeNil)

				result, err := sq.GetDirectory(id)
				So(err, ShouldBeNil)
				So(result.ClaimedBy, ShouldEqual, directory.ClaimedBy)
			})
		})
	}
}
