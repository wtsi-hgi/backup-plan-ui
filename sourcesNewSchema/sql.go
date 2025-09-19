package sourcesnewschema

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

//nolint:dupword
const insertRuleTmpl = `INSERT INTO %s
	(directoryID, backupType, backupMetadata, backupFrequency, reviewAt, deleteAt, wildcardMatch)
	VALUES (?, ?, ?, ?, DATE(?), DATE(?), ?)
`
const selectRulesTmpl = `SELECT 
    id, backupType, backupMetadata, backupFrequency, reviewAt, deleteAt, wildcardMatch
	FROM %s WHERE directoryID = ?
`
const updateRuleTmpl = `UPDATE %s SET 
	backupType = ?, backupMetadata = ?, backupFrequency = ?, reviewAt = ?, deleteAt = ? WHERE id = ?
`
const deleteTmpl = "DELETE FROM %s WHERE id = ?"

const sqlForeignKeyError = "foreign key constraint fail"

var DefaultTables = []string{DefaultDirectoriesTableName, DefaultRulesTableName} //nolint:gochecknoglobals

var ErrDirectoryDuplicate = errors.New("directory already exists")
var ErrNoDirectory = errors.New("no such directory")
var ErrNoRule = errors.New("no such rule")

type scanner interface {
	Scan(dest ...any) error
}

type SQLSourceInterface interface {
	DataSource

	ShowTables(ctx context.Context) ([]string, error)
	DropTable(ctx context.Context, tableName string) error
	DropTables(ctx context.Context) error
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

func (sq SQLSource) showTables(ctx context.Context, stmt string) ([]string, error) {
	rows, err := sq.db.QueryContext(ctx, stmt)
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

func (sq SQLSource) init(ctx context.Context, directoriesTmpl, rulesTmpl string) error {
	stmt := fmt.Sprintf(directoriesTmpl, sq.directoriesTableName)

	_, err := sq.db.ExecContext(ctx, stmt)
	if err != nil {
		return err
	}

	stmt = fmt.Sprintf(rulesTmpl, sq.rulesTableName, sq.directoriesTableName)
	_, err = sq.db.ExecContext(ctx, stmt)

	return err
}

func (sq SQLSource) DropTables(ctx context.Context) error {
	for _, tableName := range []string{sq.rulesTableName, sq.directoriesTableName} {
		err := sq.DropTable(ctx, tableName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sq SQLSource) DropTable(ctx context.Context, tableName string) error {
	stmt := fmt.Sprintf(dropTableSQLTmpl, tableName)

	_, err := sq.db.ExecContext(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to drop a table %s: %w", tableName, err)
	}

	return nil
}

func (sq SQLSource) AddDirectory(ctx context.Context, directory Directory) (id int64, err error) {
	tx, err := sq.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			callAndLogError(tx.Rollback)
		} else {
			err = tx.Commit()
		}
	}()

	id, err = sq.insertDirectory(ctx, tx, directory)
	if err != nil {
		return 0, fmt.Errorf("failed to insert directory %s: %w", directory.Path, err)
	}

	for _, rule := range directory.Rules {
		_, err = sq.setRule(ctx, tx, id, rule)
		if err != nil {
			return 0, fmt.Errorf("failed to set rule %s: %w", rule.BackupType, err)
		}
	}

	return id, nil
}

func (sq SQLSource) insertDirectory(ctx context.Context, tx *sql.Tx, directory Directory) (int64, error) {
	stmt := fmt.Sprintf(insertDirectoryTmpl, sq.directoriesTableName)

	result, err := tx.ExecContext(ctx, stmt, directory.Path, directory.Faculty, directory.Programme, directory.ClaimedBy)
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

	return insertedID, nil
}

func (sq SQLSource) setRule(ctx context.Context, tx *sql.Tx, id int64, rule Rule) (int64, error) {
	err := rule.IsValid()
	if err != nil {
		return 0, err
	}

	stmt := fmt.Sprintf(insertRuleTmpl, sq.rulesTableName)

	result, err := tx.ExecContext(ctx, stmt, id, rule.BackupType, rule.BackupMetadata, rule.BackupFrequency, rule.ReviewAt,
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

	return ruleID, nil
}

func (sq SQLSource) GetDirectory(ctx context.Context, id int64) (*Directory, error) {
	stmt := fmt.Sprintf(selectDirectoryTmpl, sq.directoriesTableName)

	row := sq.db.QueryRowContext(ctx, stmt, id)

	directory, err := sq.scanDirectory(row)
	if err != nil {
		return nil, err
	}

	directory.Rules, err = sq.GetRules(ctx, id)
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

func (sq SQLSource) SearchDirectory(ctx context.Context, path string) (*Directory, error) {
	stmt := fmt.Sprintf(selectDirectoryByPathTmpl, sq.directoriesTableName)

	row := sq.db.QueryRowContext(ctx, stmt, path)

	directory, err := sq.scanDirectory(row)
	if err != nil {
		return nil, err
	}

	directory.Rules, err = sq.GetRules(ctx, directory.ID)
	if err != nil {
		return nil, err
	}

	return directory, nil
}

func (sq SQLSource) GetRules(ctx context.Context, id int64) ([]Rule, error) {
	stmt := fmt.Sprintf(selectRulesTmpl, sq.rulesTableName)

	rows, err := sq.db.QueryContext(ctx, stmt, id)
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

func (sq SQLSource) ClaimDirectory(ctx context.Context, id int64, user string) error {
	stmt := fmt.Sprintf(claimDirectoryTmpl, sq.directoriesTableName)

	result, err := sq.db.ExecContext(ctx, stmt, user, id)
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

func (sq SQLSource) DeleteDirectory(ctx context.Context, id int64) error {
	stmt := fmt.Sprintf(deleteTmpl, sq.directoriesTableName)

	result, err := sq.db.ExecContext(ctx, stmt, id)
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

func (sq SQLSource) SetRule(ctx context.Context, id int64, rule Rule) (ruleID int64, err error) {
	tx, err := sq.db.BeginTx(ctx, nil)
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

	return sq.setRule(ctx, tx, id, rule)
}

func (sq SQLSource) DeleteRule(ctx context.Context, id int64) error {
	stmt := fmt.Sprintf(deleteTmpl, sq.rulesTableName)

	result, err := sq.db.ExecContext(ctx, stmt, id)
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

func (sq SQLSource) UpdateRule(ctx context.Context, id int64, rule Rule) error {
	stmt := fmt.Sprintf(updateRuleTmpl, sq.rulesTableName)

	result, err := sq.db.ExecContext(ctx, stmt, rule.BackupType, rule.BackupMetadata, rule.BackupFrequency,
		rule.ReviewAt, rule.DeleteAt, id)
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
