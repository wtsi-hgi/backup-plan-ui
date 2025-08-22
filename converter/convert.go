package converter

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	. "github.com/wtsi-hgi/backup-plan-ui/sources"
)

var ErrWrongEntry = errors.New("wrong entry")

func ConvertCsvToSqlite(csvPath, sqlitePath string, dropTable bool) error {
	csv := CSVSource{Path: csvPath}
	entries, err := csv.ReadAll()
	if err != nil {
		return err
	}

	for _, e := range entries {
		err = fixEntry(e)
		if err != nil {
			return err
		}
	}

	sq, err := NewSQLiteSource(sqlitePath)
	if err != nil {
		return err
	}

	defer func() {
		err = sq.Close()
		if err != nil {
			slog.Error("Failed to close SQLite connection: " + err.Error())
		}
	}()

	if dropTable {
		err = sq.DropTable()
		if err != nil {
			return err
		}
	}

	err = sq.CreateTable()
	if err != nil {
		return err
	}

	return sq.WriteEntries(entries)
}

func fixEntry(e *Entry) error {
	e.Instruction = Instruction(strings.Trim(string(e.Instruction), " "))

	if e.Instruction != Backup && e.Instruction != NoBackup && e.Instruction != TempBackup {
		return fmt.Errorf("%w: invalid instruction for entry %+v", ErrWrongEntry, e)
	}

	e.Match = strings.Trim(e.Match, " ")
	e.Ignore = strings.Trim(e.Ignore, " ")
	e.Requestor = strings.Trim(e.Requestor, " ")
	e.Faculty = strings.Trim(e.Faculty, " ")

	return nil
}

func ConvertCsvToMySQL(csvPath, tableName string, dropTable bool) error {
	csv := CSVSource{Path: csvPath}
	entries, err := csv.ReadAll()
	if err != nil {
		return err
	}

	for _, e := range entries {
		err = fixEntry(e)
		if err != nil {
			return err
		}
	}

	sq, err := NewMySQLSourceFromEnv(tableName)
	if err != nil {
		return err
	}

	defer func() {
		err = sq.Close()
		if err != nil {
			slog.Error("Failed to close MySQL connection: " + err.Error())
		}
	}()

	if dropTable {
		err = sq.DropTable()
		if err != nil {
			return err
		}
	}

	err = sq.CreateTable()
	if err != nil {
		return err
	}

	return sq.WriteEntries(entries)
}
