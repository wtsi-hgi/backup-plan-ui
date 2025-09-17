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
const claimDirectoryTmpl = "UPDATE %s SET claimedBy = ? WHERE id = ?"
const insertRuleTmpl = `INSERT INTO %s
	(directoryID, backupType, backupMetadata, backupFrequency, reviewAt, deleteAt, wildcardMatch)
	VALUES (?, ?, ?, ?, DATE(?), DATE(?), ?)
`
const selectRulesTmpl = `SELECT 
    id, backupType, backupMetadata, backupFrequency, reviewAt, deleteAt, wildcardMatch
	FROM %s WHERE directoryID = ?
`

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
	ClaimDirectory(id uint, user string) error
	SetRule(id uint, rule Rule) error
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

	directory, err := sq.scanDirectory(row)
	if err != nil {
		return nil, err
	}

	directory.Rules, err = sq.GetRules(id)
	if err != nil {
		return nil, err
	}

	return directory, nil
}

func (sq SQLSource) scanDirectory(row scanner) (*Directory, error) {
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

func (sq SQLSource) GetRules(id uint) ([]*Rule, error) {
	stmt := fmt.Sprintf(selectRulesTmpl, sq.rulesTableName)

	rows, err := sq.db.Query(stmt, id)
	if err != nil {
		return nil, err
	}

	defer callAndLogError(rows.Close)

	var rules []*Rule //nolint:prealloc

	for rows.Next() {
		rule, err := sq.scanRule(rows)
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (sq SQLSource) scanRule(row scanner) (*Rule, error) {
	var rule Rule

	err := row.Scan(&rule.ID, &rule.BackupType, &rule.BackupMetadata, &rule.BackupFrequency, &rule.ReviewAt,
		&rule.DeleteAt, &rule.WildcardMatch)
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func (sq SQLSource) ClaimDirectory(id uint, user string) error {
	stmt := fmt.Sprintf(claimDirectoryTmpl, sq.directoriesTableName)

	result, err := sq.db.Exec(stmt, user, id)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNoDirectory
	}

	return nil
}

func (sq SQLSource) SetRule(id uint, rule Rule) error {
	err := rule.IsValid()
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(insertRuleTmpl, sq.rulesTableName)

	_, err = sq.db.Exec(stmt, id, rule.BackupType, rule.BackupMetadata, rule.BackupFrequency, rule.ReviewAt,
		rule.DeleteAt, rule.WildcardMatch)
	if err != nil {
		if strings.Contains(err.Error(), "random string") {
			return ErrNoDirectory
		}

		return err
	}

	return nil
}
