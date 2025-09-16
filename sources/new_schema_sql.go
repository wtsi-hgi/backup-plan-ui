package sources

import (
	"database/sql"
	"fmt"
)

const (
	DefaultUsersTableName       = "users"
	DefaultDirectoriesTableName = "directories"
	DefaultRulesTableName       = "rules"
)

const dropTableSQLTmpl = "DROP TABLE IF EXISTS %s"

var DefaultTables = []string{DefaultUsersTableName, DefaultDirectoriesTableName, DefaultRulesTableName}

type NewSchemaSQLSource struct {
	db                   *sql.DB
	usersTableName       string
	directoriesTableName string
	rulesTableName       string
}

func (sq NewSchemaSQLSource) Close() error {
	return sq.db.Close()
}

func (sq NewSchemaSQLSource) scanTableNames(stmt string) ([]string, error) {
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

func (sq NewSchemaSQLSource) init(usersTmpl, directoriesTmpl, rulesTmpl string) error {
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

func (sq NewSchemaSQLSource) createTable(stmt string) error {
	_, err := sq.db.Exec(stmt)

	return err
}

func (sq NewSchemaMySQLSource) DropTables() error {
	for _, tableName := range []string{sq.rulesTableName, sq.directoriesTableName, sq.usersTableName} {
		err := sq.DropTable(tableName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sq NewSchemaSQLSource) DropTable(tableName string) error {
	stmt := fmt.Sprintf(dropTableSQLTmpl, tableName)
	_, err := sq.db.Exec(stmt)

	return err
}
