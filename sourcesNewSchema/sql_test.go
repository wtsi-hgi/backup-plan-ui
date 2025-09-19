package sourcesnewschema

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSQLiteSource(t *testing.T) {
	Convey("Given a database file", t, func() {
		dbFile := filepath.Join(t.TempDir(), "test.db")

		Convey("You can create a SQLite source", func() {
			sq, err := NewSQLiteSource(dbFile)
			So(err, ShouldBeNil)

			t.Cleanup(callAndLogCleanup(t, sq.Close))

			tableNames, err := sq.ShowTables(context.Background())
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

func withZeroContext(f func(context.Context) error) func() error {
	return func() error {
		return f(context.Background())
	}
}

func makeMySQLConfigForTest() MySQLConfig {
	return MySQLConfig{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     os.Getenv("MYSQL_PORT"),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASS"),
		Database: os.Getenv("MYSQL_DATABASE"),
	}
}

func TestNewMySQLSource(t *testing.T) {
	Convey("You can create a MySQL source", t, func() {
		directoriesTableName := "test_create_directories"
		rulesTableName := "test_create_rules"

		cfg := makeMySQLConfigForTest()

		err := cfg.Validate()
		if err != nil && errors.Is(err, ErrMissingArgument) {
			t.Skip("Skipping MySQL test because MySQL host, port, user, pass, or database is not set.")
		}

		sq, err := NewMySQLSource(cfg, directoriesTableName, rulesTableName)
		So(err, ShouldBeNil)

		t.Cleanup(callAndLogCleanup(t, sq.Close))
		t.Cleanup(callAndLogCleanup(t, withZeroContext(sq.DropTables)))

		tableNames, err := sq.ShowTables(context.Background())
		So(err, ShouldBeNil)

		for _, tableName := range []string{directoriesTableName, rulesTableName} {
			So(tableNames, ShouldContain, tableName)
		}
	})
}

//nolint:ireturn
func setupMySQLSourceForTest(t *testing.T) SQLSourceInterface {
	t.Helper()

	suffix := rand.Int() //nolint:gosec

	directoryTableName := fmt.Sprintf("test_directory_%d", suffix)
	ruleTableName := fmt.Sprintf("test_rule_%d", suffix)

	cfg := makeMySQLConfigForTest()

	err := cfg.Validate()
	if err != nil && errors.Is(err, ErrMissingArgument) {
		t.Skip("Skipping MySQL test because MySQL host, port, user, pass, or database is not set.")
	}

	sq, err := NewMySQLSource(cfg, directoryTableName, ruleTableName)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(callAndLogCleanup(t, sq.Close))
	t.Cleanup(callAndLogCleanup(t, withZeroContext(sq.DropTables)))

	return sq
}

//nolint:ireturn
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
	ctx := context.Background()

	var sqlTestCases = []struct {
		name  string
		setup func(t *testing.T) SQLSourceInterface
	}{
		{"SQLite", setupSQLiteSourceForTest},
		{"MySQL", setupMySQLSourceForTest},
	}

	for _, tc := range sqlTestCases {
		Convey(fmt.Sprintf("Given a %s connection", tc.name), t, func() {
			sq := tc.setup(t)

			directory := Directory{
				Path:      "/path",
				Faculty:   "test",
				Programme: "test",
			}

			Convey("You cannot get a non-existent directory", func() {
				_, err := sq.GetDirectory(ctx, 1)
				So(err, ShouldEqual, ErrNoDirectory)

				_, err = sq.SearchDirectory(ctx, "/path")
				So(err, ShouldEqual, ErrNoDirectory)
			})

			Convey("You cannot claim a non-existent directory", func() {
				err := sq.ClaimDirectory(ctx, 1, "testUser")
				So(err, ShouldEqual, ErrNoDirectory)
			})

			Convey("You cannot delete a non-existent directory", func() {
				err := sq.DeleteDirectory(ctx, 1)
				So(err, ShouldEqual, ErrNoDirectory)
			})

			Convey("You cannot update a non-existent rule", func() {
				err := sq.UpdateRule(ctx, 1, Rule{})
				So(err, ShouldEqual, ErrNoRule)
			})

			Convey("You cannot delete a non-existent rule", func() {
				err := sq.DeleteRule(ctx, 1)
				So(err, ShouldEqual, ErrNoRule)
			})

			Convey("You can add a directory", func() {
				id, err := sq.AddDirectory(ctx, directory)
				So(err, ShouldBeNil)
				So(id, ShouldBeGreaterThan, 0)

				Convey("You can get a directory", func() {
					result, err := sq.GetDirectory(ctx, id) //nolint:govet
					So(err, ShouldBeNil)

					directory.ID = id
					So(result, ShouldResemble, &directory)

					result, err = sq.SearchDirectory(ctx, directory.Path)
					So(err, ShouldBeNil)
					So(result, ShouldResemble, &directory)
				})

				Convey("You cannot add a directory with the same path", func() {
					directory2 := directory

					directory.Faculty = "test2"
					directory2.Programme = "test2"

					_, err = sq.AddDirectory(ctx, directory2)
					So(err, ShouldWrap, ErrDirectoryDuplicate)
				})

				Convey("You can claim a directory", func() {
					err = sq.ClaimDirectory(ctx, id, "testUser")
					So(err, ShouldBeNil)

					result, err := sq.GetDirectory(ctx, id) //nolint:govet
					So(err, ShouldBeNil)
					So(result.ClaimedBy, ShouldEqual, "testUser")

					Convey("You can reclaim a directory", func() {
						err = sq.ClaimDirectory(ctx, id, "testUser2")
						So(err, ShouldBeNil)

						result, err := sq.GetDirectory(ctx, id)
						So(err, ShouldBeNil)
						So(result.ClaimedBy, ShouldEqual, "testUser2")
					})
				})

				Convey("You can delete a directory", func() {
					err = sq.DeleteDirectory(ctx, id)
					So(err, ShouldBeNil)

					_, err = sq.GetDirectory(ctx, id)
					So(err, ShouldEqual, ErrNoDirectory)
				})

				Convey("And a valid rule", func() {
					rule := Rule{
						BackupType: Backup,
					}

					rule.SetDefaults()

					Convey("You can set a rule for a directory", func() {
						ruleID, err := sq.SetRule(ctx, id, rule)
						So(err, ShouldBeNil)
						So(ruleID, ShouldBeGreaterThan, 0)

						rule.ID = ruleID
						rule.ReviewAt = rule.ReviewAt.Truncate(24 * time.Hour)
						rule.DeleteAt = rule.DeleteAt.Truncate(24 * time.Hour)

						result, err := sq.GetDirectory(ctx, id)
						So(err, ShouldBeNil)
						So(result.Rules, ShouldHaveLength, 1)
						So(result.Rules[0], ShouldResemble, rule)

						Convey("You can delete a rule for a directory", func() {
							err = sq.DeleteRule(ctx, rule.ID)
							So(err, ShouldBeNil)

							result, err = sq.GetDirectory(ctx, id)
							So(err, ShouldBeNil)
							So(result.Rules, ShouldHaveLength, 0)
						})

						Convey("You can update a rule for a directory", func() {
							rule.BackupType = NoBackup

							err = sq.UpdateRule(ctx, rule.ID, rule)
							So(err, ShouldBeNil)

							result, err = sq.GetDirectory(ctx, id)
							So(err, ShouldBeNil)
							So(result.Rules, ShouldHaveLength, 1)
							So(result.Rules[0].BackupType, ShouldResemble, NoBackup)
						})
					})
				})
			})

			Convey("You can add a directory with a claimed user", func() {
				directory.ClaimedBy = "testUser"

				id, err := sq.AddDirectory(ctx, directory)
				So(err, ShouldBeNil)

				result, err := sq.GetDirectory(ctx, id)
				So(err, ShouldBeNil)
				So(result.ClaimedBy, ShouldEqual, directory.ClaimedBy)
			})

			Convey("Given a valid rule", func() {
				rule := Rule{
					BackupType: Backup,
				}

				rule.SetDefaults()

				Convey("You cannot set a rule for a non-existent directory", func() {
					_, err := sq.SetRule(ctx, 1, rule)
					So(err, ShouldEqual, ErrNoDirectory)
				})

				Convey("You can add a directory with a rule", func() {
					directory.AddRule(rule)

					id, err := sq.AddDirectory(ctx, directory)
					So(err, ShouldBeNil)

					rule.ID = 1
					rule.ReviewAt = rule.ReviewAt.Truncate(24 * time.Hour)
					rule.DeleteAt = rule.DeleteAt.Truncate(24 * time.Hour)

					result, err := sq.GetDirectory(ctx, id)
					So(err, ShouldBeNil)
					So(result.Rules, ShouldHaveLength, 1)
					So(result.Rules[0], ShouldResemble, rule)
				})
			})
		})
	}
}
