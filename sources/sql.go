package sources

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type SQLSource struct {
	db        *sql.DB
	tableName string
}

type SQLiteSource struct {
	*SQLSource
}

type MySQLSource struct {
	*SQLSource
}

const defaultTableName = "entries"

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
	getAllStmt          = "SELECT * FROM %s"
	getEntryStmt        = "SELECT * FROM %s WHERE id = ?"
	deleteEntryStmt     = "DELETE FROM %s WHERE id = ?"
	deleteReturningStmt = "DELETE FROM %s WHERE id = ? RETURNING *"
	updateEntryStmt     = `UPDATE %s 
					   SET reporting_name = ?, reporting_root = ?, directory = ?, instruction = ?, 
                       keep = ?, skip = ?, requestor = ?, faculty = ? WHERE id = ?`
	insertEntryStmt = `INSERT INTO %s 
			          (reporting_name, reporting_root, directory, instruction, keep, skip, requestor, faculty) 
			          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
)

func (sq SQLSource) callAndLogError(f func() error) {
	err := f()
	if err != nil {
		slog.Error(err.Error())
	}
}

// NewSQLiteSource opens a connection to an SQLite database at the given path and stores it internally.
// You are responsible to close the connection using Close().
func NewSQLiteSource(path string) (SQLiteSource, error) {
	db, err := sql.Open("sqlite3", path)

	return SQLiteSource{&SQLSource{db: db, tableName: defaultTableName}}, err
}

// NewMySQLSource opens a connection to a MySQL database using given credentials and stores it internally.
// You are responsible to close the connection using Close().
func NewMySQLSource(host, port, user, password, dbName, tableName string) (MySQLSource, error) {
	address := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	db, err := sql.Open("mysql", address)

	return MySQLSource{&SQLSource{db: db, tableName: tableName}}, err
}

func (sq SQLSource) Close() error {
	return sq.db.Close()
}

func (sq SQLiteSource) CreateTable() error {
	createTableStmt := fmt.Sprintf(createTableTmpl, sq.tableName, "AUTOINCREMENT", Backup, NoBackup, TempBackup)
	_, err := sq.db.Exec(createTableStmt)

	return err
}

func (sq MySQLSource) CreateTable() error {
	createTableStmt := fmt.Sprintf(createTableTmpl, sq.tableName, "AUTO_INCREMENT", Backup, NoBackup, TempBackup)
	_, err := sq.db.Exec(createTableStmt)

	return err
}

func (sq SQLSource) ReadAll() ([]*Entry, error) {
	rows, err := sq.db.Query(fmt.Sprintf(getAllStmt, sq.tableName))
	if err != nil {
		return nil, err
	}

	defer sq.callAndLogError(rows.Close)

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
	stmt := fmt.Sprintf(getEntryStmt, sq.tableName)

	row := sq.db.QueryRow(stmt, id)

	entry, err := sq.scanEntry(row)

	return entry, err
}

func (sq SQLSource) UpdateEntry(newEntry *Entry) error {
	stmt := fmt.Sprintf(updateEntryStmt, sq.tableName)

	r, err := sq.db.Exec(stmt, newEntry.ReportingName, newEntry.ReportingRoot, newEntry.Directory,
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

func (sq SQLiteSource) DeleteEntry(id uint16) (*Entry, error) {
	stmt := fmt.Sprintf(deleteReturningStmt, sq.tableName)

	row := sq.db.QueryRow(stmt, id)

	entry, err := sq.scanEntry(row)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoEntry
	}

	return entry, err
}

func (sq MySQLSource) DeleteEntry(id uint16) (*Entry, error) {
	tx, err := sq.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			sq.callAndLogError(tx.Rollback)
		} else {
			err = tx.Commit()
		}
	}()

	getStmt := fmt.Sprintf(getEntryStmt, sq.tableName)
	row := tx.QueryRow(getStmt, id)

	entry, err := sq.scanEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoEntry
		}

		return nil, err
	}

	delStmt := fmt.Sprintf(deleteEntryStmt, sq.tableName)

	_, err = tx.Exec(delStmt, id)
	if err != nil {
		return nil, err
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

	defer func() {
		if err != nil {
			sq.callAndLogError(tx.Rollback)
		} else {
			err = tx.Commit()
		}
	}()

	stmt, err := tx.Prepare(fmt.Sprintf(insertEntryStmt, sq.tableName))
	if err != nil {
		return err
	}
	defer sq.callAndLogError(stmt.Close)

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

	return err
}
