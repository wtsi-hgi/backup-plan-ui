package converter

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/backup-plan-ui/sources"
	"github.com/wtsi-hgi/backup-plan-ui/sourcesNewSchema"
)

func TestConvertEntry(t *testing.T) {
	Convey("Given a minimal test entry", t, func() {
		entry := sources.Entry{
			ReportingName: "Test Project",
			ReportingRoot: "/path",
			Directory:     "/path/project",
			Instruction:   sources.Backup,
			Faculty:       "Test Faculty",
			Requestor:     "Test User",
		}

		Convey("You can convert it to a directory", func() {
			directory, err := convertEntry(&entry)
			So(err, ShouldBeNil)

			So(directory.Path, ShouldEqual, entry.Directory)
			So(directory.Faculty, ShouldEqual, entry.Faculty)
			So(directory.ClaimedBy, ShouldEqual, entry.Requestor)

			Convey("With a default rule", func() {
				So(directory.Rules, ShouldHaveLength, 1)

				So(directory.Rules[0].BackupType, ShouldEqual, string(entry.Instruction))
				So(directory.Rules[0].WildcardMatch, ShouldEqual, "*")
			})
		})

		Convey("With two match patterns", func() {
			entry.Match = "*.sh *.txt"

			Convey("Convert will make two rules", func() {
				directory, err := convertEntry(&entry)
				So(err, ShouldBeNil)

				So(directory.Rules, ShouldHaveLength, 2)

				So(directory.Rules[0].BackupType, ShouldEqual, string(entry.Instruction))
				So(directory.Rules[0].WildcardMatch, ShouldEqual, "*.sh")

				So(directory.Rules[1].BackupType, ShouldEqual, string(entry.Instruction))
				So(directory.Rules[1].WildcardMatch, ShouldEqual, "*.txt")
			})
		})

		Convey("With two ignore patterns", func() {
			entry.Ignore = "*.sh *.txt"

			Convey("Convert will make three rules", func() {
				directory, err := convertEntry(&entry)
				So(err, ShouldBeNil)

				So(directory.Rules, ShouldHaveLength, 3)

				So(directory.Rules[0].BackupType, ShouldEqual, string(entry.Instruction))
				So(directory.Rules[0].WildcardMatch, ShouldEqual, "*")

				So(directory.Rules[1].BackupType, ShouldEqual, string(sources.NoBackup))
				So(directory.Rules[1].WildcardMatch, ShouldEqual, "*.sh")

				So(directory.Rules[2].BackupType, ShouldEqual, string(sources.NoBackup))
				So(directory.Rules[2].WildcardMatch, ShouldEqual, "*.txt")
			})
		})
	})
}

func TestConvertSchema(t *testing.T) {
	Convey("Given some test data", t, func() {
		entriesTable := "test_convert_entries"
		db, err := sources.NewMySQLSourceFromEnv(entriesTable)
		if err != nil && errors.Is(err, sources.ErrMissingArgument) {
			t.Skip("Skipping MySQL test because MySQL host, port, user, pass, or database is not set.")
		}

		So(err, ShouldBeNil)

		t.Cleanup(callAndLogCleanup(t, db.Close))
		t.Cleanup(callAndLogCleanup(t, db.DropTable))

		testEntries := sources.CreateTestEntries(t)

		extraEntry := *testEntries[0]
		extraEntry.Instruction = sources.TempBackup
		extraEntry.Directory = "/tmp"
		extraEntry.Ignore = "*.log"

		testEntries = append(testEntries, &extraEntry)

		err = db.WriteEntries(testEntries)
		So(err, ShouldBeNil)

		Convey("You can convert schema", func() {
			dirsTable := "test_convert_dirs"
			rulesTable := "test_convert_rules"

			err = ConvertSchema(entriesTable, dirsTable, rulesTable, false)
			So(err, ShouldBeNil)

			newDB, err := sourcesNewSchema.NewMySQLSourceFromEnv(dirsTable, rulesTable)
			So(err, ShouldBeNil)

			t.Cleanup(callAndLogCleanup(t, newDB.Close))
			t.Cleanup(callAndLogCleanup(t, newDB.DropTables))

			for _, entry := range testEntries {
				Convey(fmt.Sprintf("%s - %s", entry.Directory, entry.Instruction), func() {
					directory, err := newDB.SearchDirectory(entry.Directory)
					So(err, ShouldBeNil)

					compareEntryAndDirectory(t, entry, directory)
				})
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

func compareEntryAndDirectory(t *testing.T, entry *sources.Entry, directory *sourcesNewSchema.Directory) {
	t.Helper()

	So(entry.Directory, ShouldEqual, directory.Path)
	So(entry.Faculty, ShouldEqual, directory.Faculty)
	So(entry.Requestor, ShouldEqual, directory.ClaimedBy)

	So(isInstructionInRules(entry.Instruction, entry.Match, entry.Ignore, directory.Rules), ShouldBeTrue)
}

func isInstructionInRules(instruction sources.Instruction, match string, ignore string, rules []sourcesNewSchema.Rule) bool {
	matchPresent := false
	ignorePresent := true

	for _, rule := range rules {
		if ruleResemblesInstruction(instruction, match, rule) {
			matchPresent = true

			break
		}
	}

	if ignore != "" {
		ignorePresent = isInstructionInRules(sources.NoBackup, ignore, "", rules)
	}

	return matchPresent && ignorePresent
}

func ruleResemblesInstruction(instruction sources.Instruction, match string, rule sourcesNewSchema.Rule) bool {
	if rule.BackupType != string(instruction) {
		return false
	}

	if match != "" && rule.WildcardMatch != match {
		return false
	}

	return true
}
