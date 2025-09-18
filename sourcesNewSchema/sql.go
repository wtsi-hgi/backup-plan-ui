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
const selectDirectoryByPathTmpl = "SELECT id, path, faculty, programme, claimedBy FROM %s WHERE path = ?"
const claimDirectoryTmpl = "UPDATE %s SET claimedBy = ? WHERE id = ?"
const insertRuleTmpl = `INSERT INTO %s
	(directoryID, backupType, backupMetadata, backupFrequency, reviewAt, deleteAt, wildcardMatch)
	VALUES (?, ?, ?, ?, DATE(?), DATE(?), ?)
`
const selectRulesTmpl = `SELECT 
    id, backupType, backupMetadata, backupFrequency, reviewAt, deleteAt, wildcardMatch
	FROM %s WHERE directoryID = ?
`
const updateRuleTmpl = "UPDATE %s SET backupType = ?, backupMetadata = ?, backupFrequency = ?, reviewAt = ?, deleteAt = ? WHERE id = ?"
const deleteTmpl = "DELETE FROM %s WHERE id = ?"

const sqlForeignKeyError = "foreign key constraint fail"

var DefaultTables = []string{DefaultDirectoriesTableName, DefaultRulesTableName}

var ErrDirectoryDuplicate = errors.New("directory already exists")
var ErrNoDirectory = errors.New("no such directory")
var ErrNoRule = errors.New("no such rule")

type scanner interface {
	Scan(dest ...any) error
}

type SQLSourceInterface interface {
	DataSource

	ShowTables() ([]string, error)
	DropTable(tableName string) error
	DropTables() error
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

func (sq SQLSource) AddDirectory(directory Directory) (id uint, err error) {
	tx, err := sq.db.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			callAndLogError(tx.Rollback)
		} else {
			err = tx.Commit()
		}
	}()

	id, err = sq.insertDirectory(tx, directory)
	if err != nil {
		return 0, err
	}

	for _, rule := range directory.Rules {
		_, err = sq.setRule(tx, id, rule)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (sq SQLSource) insertDirectory(tx *sql.Tx, directory Directory) (uint, error) {
	stmt := fmt.Sprintf(insertDirectoryTmpl, sq.directoriesTableName)

	result, err := tx.Exec(stmt, directory.Path, directory.Faculty, directory.Programme, directory.ClaimedBy)
	if err != nil {
		if strings.Contains(err.Error(), sq.dupRowsError) {
			return 0, ErrDirectoryDuplicate
		}

		return 0, err
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint(insertedID), nil
}

func (sq SQLSource) setRule(tx *sql.Tx, id uint, rule Rule) (uint, error) {
	err := rule.IsValid()
	if err != nil {
		return 0, err
	}

	stmt := fmt.Sprintf(insertRuleTmpl, sq.rulesTableName)

	result, err := tx.Exec(stmt, id, rule.BackupType, rule.BackupMetadata, rule.BackupFrequency, rule.ReviewAt,
		rule.DeleteAt, rule.WildcardMatch)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), sqlForeignKeyError) {
			return 0, ErrNoDirectory
		}

		return 0, err
	}

	ruleID, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return uint(ruleID), nil
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

func (sq SQLSource) SearchDirectory(path string) (*Directory, error) {
	stmt := fmt.Sprintf(selectDirectoryByPathTmpl, sq.directoriesTableName)

	row := sq.db.QueryRow(stmt, path)

	directory, err := sq.scanDirectory(row)
	if err != nil {
		return nil, err
	}

	directory.Rules, err = sq.GetRules(directory.ID)
	if err != nil {
		return nil, err
	}

	return directory, nil
}

func (sq SQLSource) GetRules(id uint) ([]Rule, error) {
	stmt := fmt.Sprintf(selectRulesTmpl, sq.rulesTableName)

	rows, err := sq.db.Query(stmt, id)
	if err != nil {
		return nil, err
	}

	defer callAndLogError(rows.Close)

	var rules []Rule //nolint:prealloc

	for rows.Next() {
		rule, err := sq.scanRule(rows)
		if err != nil {
			return nil, err
		}

		rules = append(rules, *rule)
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

func (sq SQLSource) DeleteDirectory(id uint) error {
	stmt := fmt.Sprintf(deleteTmpl, sq.directoriesTableName)

	result, err := sq.db.Exec(stmt, id)
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

func (sq SQLSource) SetRule(id uint, rule Rule) (ruleID uint, err error) {
	tx, err := sq.db.Begin()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			callAndLogError(tx.Rollback)
		} else {
			err = tx.Commit()
		}
	}()

	return sq.setRule(tx, id, rule)
}

func (sq SQLSource) DeleteRule(id uint) error {
	stmt := fmt.Sprintf(deleteTmpl, sq.rulesTableName)
	result, err := sq.db.Exec(stmt, id)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNoRule
	}

	return nil
}

func (sq SQLSource) UpdateRule(id uint, rule Rule) error {
	stmt := fmt.Sprintf(updateRuleTmpl, sq.rulesTableName)

	result, err := sq.db.Exec(stmt, rule.BackupType, rule.BackupMetadata, rule.BackupFrequency, rule.ReviewAt,
		rule.DeleteAt, id)

	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNoRule
	}

	return nil
}
