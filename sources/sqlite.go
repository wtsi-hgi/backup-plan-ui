package sources

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteSource struct {
	db *sql.DB
}

var createTableStmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS entries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	reporting_name TEXT,
	reporting_root TEXT,
	directory TEXT,
	instruction TEXT CHECK ( instruction IN ('%s', '%s', '%s') ),
	match TEXT,
	ignore TEXT,
	requestor TEXT,
	faculty TEXT
)`, Backup, NoBackup, TempBackup)

const (
	getAllStmt      = "SELECT * FROM entries"
	getEntryStmt    = "SELECT * FROM entries WHERE id = ?"
	deleteEntryStmt = "DELETE FROM entries WHERE id = ? RETURNING *"
	updateEntryStmt = `UPDATE entries 
					   SET reporting_name = ?, reporting_root = ?, directory = ?, instruction = ?, 
                       match = ?, ignore = ?, requestor = ?, faculty = ? WHERE id = ?`
	insertEntryStmt = `INSERT INTO entries 
			          (reporting_name, reporting_root, directory, instruction, match, ignore, requestor, faculty) 
			          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
)

// NewSQLiteSource opens a connection and stores it internally.
// You are responsible to close it using Close().
func NewSQLiteSource(path string) (SQLiteSource, error) {
	db, err := sql.Open("sqlite3", path)

	return SQLiteSource{db: db}, err
}

func (sq SQLiteSource) Close() error {
	return sq.db.Close()
}

func (sq SQLiteSource) CreateTable() error {
	_, err := sq.db.Exec(createTableStmt)

	return err
}

func (sq SQLiteSource) ReadAll() ([]*Entry, error) {
	rows, err := sq.db.Query(getAllStmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var entries []*Entry

	for rows.Next() {
		entry, err := sq.scanEntry(rows)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (sq SQLiteSource) scanEntry(row scanner) (*Entry, error) {
	var entry Entry

	err := row.Scan(&entry.ID, &entry.ReportingName, &entry.ReportingRoot, &entry.Directory,
		&entry.Instruction, &entry.Match, &entry.Ignore, &entry.Requestor, &entry.Faculty)

	return &entry, err
}

func (sq SQLiteSource) GetEntry(id uint16) (*Entry, error) {
	row := sq.db.QueryRow(getEntryStmt, id)

	entry, err := sq.scanEntry(row)

	return entry, err
}

func (sq SQLiteSource) UpdateEntry(newEntry *Entry) error {
	_, err := sq.db.Exec(updateEntryStmt, newEntry.ReportingName, newEntry.ReportingRoot, newEntry.Directory,
		newEntry.Instruction, newEntry.Match, newEntry.Ignore, newEntry.Requestor, newEntry.Faculty, newEntry.ID)

	return err
}

func (sq SQLiteSource) DeleteEntry(id uint16) (*Entry, error) {
	row := sq.db.QueryRow(deleteEntryStmt, id)

	entry, err := sq.scanEntry(row)

	return entry, err
}

func (sq SQLiteSource) AddEntry(entry *Entry) error {
	return sq.writeEntries([]*Entry{entry})
}

func (sq SQLiteSource) writeEntries(entries []*Entry) error {
	tx, err := sq.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(insertEntryStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err = stmt.Exec(entry.ReportingName, entry.ReportingRoot, entry.Directory,
			entry.Instruction, entry.Match, entry.Ignore, entry.Requestor, entry.Faculty)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
