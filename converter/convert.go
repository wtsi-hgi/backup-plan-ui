package converter

import (
	"errors"
	"fmt"
	"log/slog"

	. "backup-plan-ui/sources"
)

func ConvertCsvToSqlite(csvPath, sqlitePath string) error {
	csv := CSVSource{Path: csvPath}
	entries, err := csv.ReadAll()
	if err != nil {
		return err
	}

	for i, e := range entries {
		if e.Instruction != Backup && e.Instruction != NoBackup && e.Instruction != TempBackup {
			fmt.Printf("Entry %d: %+v\n", i, e)
			return errors.New("wrong entry")
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

	err = sq.CreateTable()
	if err != nil {
		return err
	}

	return sq.WriteEntries(entries)
}
