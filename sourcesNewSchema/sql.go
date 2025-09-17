package sourcesNewSchema

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DefaultDirectoriesTableName = "directories"
	DefaultRulesTableName       = "rules"
)

const dropTableSQLTmpl = "DROP TABLE IF EXISTS %s"

const insertDirectoryTmpl = "INSERT INTO %s (path, faculty, programme, claimedBy) VALUES (?, ?, ?, ?)"
const selectDirectoryTmpl = "SELECT id, path, faculty, programme, claimedBy FROM %s WHERE id = ?"

var DefaultTables = []string{DefaultDirectoriesTableName, DefaultRulesTableName}

var ErrDirectoryDuplicate = errors.New("directory already exists")
var ErrNoDirectory = errors.New("no such directory")

type scanner interface {
	Scan(dest ...any) error
}

type SQLSourceInterface interface {
	Init() error
	Close() error
	ShowTables() ([]string, error)
	DropTable(tableName string) error
	DropTables() error
	AddDirectory(directory Directory) (uint, error)
	GetDirectory(id uint) (*Directory, error)
}

// SQLSource is a type for shared functionality between MySQL and SQLite.
type SQLSource struct {
	db                   *sql.DB
	directoriesTableName string
	rulesTableName       string
	dupRowsError         string
}

func (sq SQLSource) Close() error {
	return sq.db.Close()
}

func (sq SQLSource) showTables(stmt string) ([]string, error) {
	rows, err := sq.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer callAndLogError(rows.Close)

	var tableName string
	tableNames := make([]string, 0, len(DefaultTables))

	for rows.Next() {
		err = rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}

		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

func callAndLogError(f func() error) {
	err := f()
	if err != nil {
		slog.Error(err.Error())
	}
}

func (sq SQLSource) init(directoriesTmpl, rulesTmpl string) error {
	stmt := fmt.Sprintf(directoriesTmpl, sq.directoriesTableName)
	_, err := sq.db.Exec(stmt)
	if err != nil {
		return err
	}

	stmt = fmt.Sprintf(rulesTmpl, sq.rulesTableName, sq.directoriesTableName)
	_, err = sq.db.Exec(stmt)

	return err
}

func (sq SQLSource) createTable(stmt string) error {
	_, err := sq.db.Exec(stmt)

	return err
}

func (sq SQLSource) DropTables() error {
	for _, tableName := range []string{sq.rulesTableName, sq.directoriesTableName} {
		err := sq.DropTable(tableName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sq SQLSource) DropTable(tableName string) error {
	stmt := fmt.Sprintf(dropTableSQLTmpl, tableName)
	_, err := sq.db.Exec(stmt)

	return err
}

func (sq SQLSource) AddDirectory(directory Directory) (uint, error) {
	stmt := fmt.Sprintf(insertDirectoryTmpl, sq.directoriesTableName)

	result, err := sq.db.Exec(stmt, directory.Path, directory.Faculty, directory.Programme, directory.ClaimedBy)
	if err != nil {
		if strings.Contains(err.Error(), sq.dupRowsError) {
			return 0, ErrDirectoryDuplicate
		}

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func (sq SQLSource) GetDirectory(id uint) (*Directory, error) {
	stmt := fmt.Sprintf(selectDirectoryTmpl, sq.directoriesTableName)

	row := sq.db.QueryRow(stmt, id)

	var directory Directory

	err := row.Scan(&directory.ID, &directory.Path, &directory.Faculty, &directory.Programme, &directory.ClaimedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoDirectory
		}

		return nil, err
	}

	return &directory, nil
}
