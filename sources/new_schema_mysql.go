package sources

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
)

const createMySQLUsersTableTmpl = `CREATE TABLE IF NOT EXISTS %s (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    userName VARCHAR(10) NOT NULL,
    faculty VARCHAR(30) NOT NULL,
    programme VARCHAR(30) NOT NULL
)`

const createMySQLDirectoriesTableTmpl = `CREATE TABLE IF NOT EXISTS %s (
	id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
	path TEXT NOT NULL,
	claimedByUserID INT UNSIGNED,
	
	FOREIGN KEY (claimedByUserID) REFERENCES %s(id) ON DELETE SET NULL
)`

const createMySQLRulesTableTmpl = `CREATE TABLE IF NOT EXISTS %s (
	id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
	directoryID INT UNSIGNED NOT NULL,
	backupType TINYTEXT NOT NULL,
	backupMetadata TEXT,
	backupFrequency SMALLINT UNSIGNED NOT NULL,
	reviewAt DATE,
	deleteAt DATE,
	wildcardMatch TEXT,
	
	FOREIGN KEY (directoryID) REFERENCES %s(id) ON DELETE CASCADE
)`

type NewSchemaMySQLSource struct {
	*NewSchemaSQLSource
}

func NewSchemaNewMySQLSourceFromEnv(usersTableName, directoriesTableName, rulesTableName string,
) (*NewSchemaMySQLSource, error) {
	return NewSchemaNewMySQLSource(
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_DATABASE"),
		usersTableName,
		directoriesTableName,
		rulesTableName,
	)
}

// NewSchemaNewMySQLSource opens a connection to a MySQL database using given credentials and stores it internally.
// It also creates a table with the given name if it does not exist.
// You are responsible to close the connection using Close().
func NewSchemaNewMySQLSource(host, port, user, password, dbName, usersTableName, directoriesTableName,
	rulesTableName string) (*NewSchemaMySQLSource, error) {
	var missing []string

	appendIfEmpty(&missing, "host", host)
	appendIfEmpty(&missing, "port", port)
	appendIfEmpty(&missing, "user", user)
	appendIfEmpty(&missing, "password", password)
	appendIfEmpty(&missing, "dbName", dbName)

	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: %v\n", ErrMissingArgument, missing)
	}

	address := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	db, err := sql.Open("mysql", address)
	if err != nil {
		return nil, err
	}

	sq := &NewSchemaMySQLSource{
		&NewSchemaSQLSource{
			db:                   db,
			usersTableName:       usersTableName,
			directoriesTableName: directoriesTableName,
			rulesTableName:       rulesTableName,
		},
	}

	return sq, sq.Init()
}

func appendIfEmpty(array *[]string, name, val string) {
	if val == "" {
		*array = append(*array, name)
	}
}

func (sq NewSchemaMySQLSource) ShowTables() ([]string, error) {
	return sq.scanTableNames("SHOW TABLES")
}

func (sq NewSchemaMySQLSource) Init() error {
	tables, err := sq.ShowTables()
	if err != nil {
		return err
	}

	initialised := true
	requiredTables := []string{sq.usersTableName, sq.directoriesTableName, sq.rulesTableName}

	for _, requiredTable := range requiredTables {
		if !slices.Contains(tables, requiredTable) {
			initialised = false

			break
		}
	}

	if initialised {
		return nil
	}

	return sq.init(
		createMySQLUsersTableTmpl,
		createMySQLDirectoriesTableTmpl,
		createMySQLRulesTableTmpl,
	)
}
