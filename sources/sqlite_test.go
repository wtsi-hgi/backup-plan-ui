package sources

import (
	"log"
	"path/filepath"
	"testing"

	. "github.com/smarty/assertions"
)

func TestSQLiteSource_ReadAll(t *testing.T) {
	entries, sq := createTestTable(t)
	defer sq.Close()

	testDataSourceReadAll(t, sq, entries)
}

func TestSQLiteSource_GetEntry(t *testing.T) {
	entries, sq := createTestTable(t)
	defer sq.Close()

	testDataSourceGetEntry(t, sq, entries)
}

func TestSQLiteSource_UpdateEntry(t *testing.T) {
	entries, sq := createTestTable(t)
	defer sq.Close()

	testDataSourceUpdateEntry(t, sq, entries)
}

func TestSQLiteSource_DeleteEntry(t *testing.T) {
	testCases := []struct {
		name    string
		entryID uint16
	}{
		{"Delete first entry", 1},
		{"Delete middle entry", max(1, NumTestDataRows-1)},
		{"Delete last entry", NumTestDataRows},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			entries, sq := createTestTable(t)
			defer sq.Close()

			testDataSourceDeleteEntry(t, sq, entries[tt.entryID-1], tt.entryID)
		})
	}
}

func TestSQLiteSource_AddEntry(t *testing.T) {
	entries, sq := createTestTable(t)
	defer sq.Close()

	testDataSourceAddEntry(t, sq, entries)
}

func TestSQLiteSource_WriteEntries(t *testing.T) {
	entries := CreateTestEntries(t)

	dbPath := filepath.Join(t.TempDir(), "test.db")

	sq, err := NewSQLiteSource(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer sq.Close()

	err = sq.CreateTable()
	if err != nil {
		t.Fatal(err)
	}

	err = sq.writeEntries(entries)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateTable(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "test.db")

	sq, err := NewSQLiteSource(dbFile)
	if err != nil {
		t.Fatal(err)
	}
	defer sq.Close()

	err = sq.CreateTable()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := sq.db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tableName string
	var tableNames []string

	for rows.Next() {
		err = rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}

		tableNames = append(tableNames, tableName)
	}

	if ok, err := So(tableNames, ShouldContain, entriesTableName); !ok {
		log.Fatal(err)
	}
}
