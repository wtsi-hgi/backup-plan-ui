package sources

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type SQLSource struct {
	db *sql.DB
}

type SQLiteSource struct {
	*SQLSource
}

type MySQLSource struct {
	*SQLSource
}

const createTableTmpl = `CREATE TABLE IF NOT EXISTS %s (
	id INTEGER PRIMARY KEY %s,
	reporting_name TEXT,
	reporting_root TEXT,
	directory TEXT,
	instruction TEXT CHECK ( instruction IN ('%s', '%s', '%s') ),
	keep TEXT,
	skip TEXT,
	requestor TEXT,
	faculty TEXT
)`

const (
	getAllStmt      = "SELECT * FROM entries"
	getEntryStmt    = "SELECT * FROM entries WHERE id = ?"
	deleteEntryStmt = "DELETE FROM entries WHERE id = ? RETURNING *"
	updateEntryStmt = `UPDATE entries 
					   SET reporting_name = ?, reporting_root = ?, directory = ?, instruction = ?, 
                       keep = ?, skip = ?, requestor = ?, faculty = ? WHERE id = ?`
	insertEntryStmt = `INSERT INTO entries 
			          (reporting_name, reporting_root, directory, instruction, keep, skip, requestor, faculty) 
			          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
)

// NewSQLiteSource opens a connection to an SQLite database at the given path and stores it internally.
// You are responsible to close the connection using Close().
func NewSQLiteSource(path string) (SQLiteSource, error) {
	db, err := sql.Open("sqlite3", path)

	return SQLiteSource{&SQLSource{db: db}}, err
}

// NewMySQLSource opens a connection to a MySQL database using given credentials and stores it internally.
// You are responsible to close the connection using Close().
func NewMySQLSource(host, port, user, password, dbName string) (MySQLSource, error) {
	address := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	db, err := sql.Open("mysql", address)

	return MySQLSource{&SQLSource{db: db}}, err
}

func (sq SQLSource) Close() error {
	return sq.db.Close()
}

func (sq SQLiteSource) CreateTable() error {
	createTableStmt := fmt.Sprintf(createTableTmpl, "entries", "AUTOINCREMENT", Backup, NoBackup, TempBackup)
	_, err := sq.db.Exec(createTableStmt)

	return err
}

func (sq MySQLSource) CreateTable(name string) error {
	createTableStmt := fmt.Sprintf(createTableTmpl, name, "AUTO_INCREMENT", Backup, NoBackup, TempBackup)
	_, err := sq.db.Exec(createTableStmt)

	return err
}

func (sq SQLSource) ReadAll() ([]*Entry, error) {
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

func (sq SQLSource) scanEntry(row scanner) (*Entry, error) {
	var entry Entry

	err := row.Scan(&entry.ID, &entry.ReportingName, &entry.ReportingRoot, &entry.Directory,
		&entry.Instruction, &entry.Match, &entry.Ignore, &entry.Requestor, &entry.Faculty)

	return &entry, err
}

func (sq SQLSource) GetEntry(id uint16) (*Entry, error) {
	row := sq.db.QueryRow(getEntryStmt, id)

	entry, err := sq.scanEntry(row)

	return entry, err
}

func (sq SQLSource) UpdateEntry(newEntry *Entry) error {
	r, err := sq.db.Exec(updateEntryStmt, newEntry.ReportingName, newEntry.ReportingRoot, newEntry.Directory,
		newEntry.Instruction, newEntry.Match, newEntry.Ignore, newEntry.Requestor, newEntry.Faculty, newEntry.ID)

	if err != nil {
		return err
	}

	count, err := r.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoEntry
	}

	return nil
}

func (sq SQLSource) DeleteEntry(id uint16) (*Entry, error) {
	row := sq.db.QueryRow(deleteEntryStmt, id)

	entry, err := sq.scanEntry(row)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoEntry
	}

	return entry, err
}

func (sq SQLSource) AddEntry(entry *Entry) error {
	return sq.WriteEntries([]*Entry{entry})
}

func (sq SQLSource) WriteEntries(entries []*Entry) error {
	tx, err := sq.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertEntryStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		r, err := stmt.Exec(entry.ReportingName, entry.ReportingRoot, entry.Directory,
			entry.Instruction, entry.Match, entry.Ignore, entry.Requestor, entry.Faculty)

		if err != nil {
			return err
		}

		id, err := r.LastInsertId()
		if err != nil {
			return err
		}

		entry.ID = uint16(id)
	}

	return tx.Commit()
}
