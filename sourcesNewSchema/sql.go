package sourcesNewSchema

import (
	"database/sql"
	"fmt"
	"log/slog"
)

const (
	DefaultUsersTableName       = "users"
	DefaultDirectoriesTableName = "directories"
	DefaultRulesTableName       = "rules"
)

const dropTableSQLTmpl = "DROP TABLE IF EXISTS %s"

const insertUserTmpl = "INSERT INTO %s (userName, faculty, programme) VALUES (?, ?, ?)"

const selectUserTmpl = "SELECT id, userName, faculty, programme FROM %s WHERE id = ?"

var DefaultTables = []string{DefaultUsersTableName, DefaultDirectoriesTableName, DefaultRulesTableName}

type scanner interface {
	Scan(dest ...any) error
}

type SQLSourceInterface interface {
	Init() error
	Close() error
	ShowTables() ([]string, error)
	DropTable(tableName string) error
	DropTables() error
	AddUser(user User) (uint16, error)
	GetUser(id uint16) (*User, error)
}

// SQLSource is a type for shared functionality between MySQL and SQLite.
type SQLSource struct {
	db                   *sql.DB
	usersTableName       string
	directoriesTableName string
	rulesTableName       string
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

func (sq SQLSource) init(usersTmpl, directoriesTmpl, rulesTmpl string) error {
	stmt := fmt.Sprintf(usersTmpl, sq.usersTableName)
	_, err := sq.db.Exec(stmt)
	if err != nil {
		return err
	}

	stmt = fmt.Sprintf(directoriesTmpl, sq.directoriesTableName, sq.usersTableName)
	_, err = sq.db.Exec(stmt)
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
	for _, tableName := range []string{sq.rulesTableName, sq.directoriesTableName, sq.usersTableName} {
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

func (sq SQLSource) AddUser(user User) (uint16, error) {
	stmt := fmt.Sprintf(insertUserTmpl, sq.usersTableName)

	result, err := sq.db.Exec(stmt, user.Username, user.Faculty, user.Programme)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint16(id), nil
}

func (sq SQLSource) GetUser(id uint16) (*User, error) {
	stmt := fmt.Sprintf(selectUserTmpl, sq.usersTableName)

	row := sq.db.QueryRow(stmt, id)

	var user User

	err := row.Scan(&user.ID, &user.Username, &user.Faculty, &user.Programme)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
