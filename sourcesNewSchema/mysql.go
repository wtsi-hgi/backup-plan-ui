package sourcesNewSchema

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"slices"
)

const createMySQLDirectoriesTableTmpl = `CREATE TABLE IF NOT EXISTS %s (
	id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
	path TEXT NOT NULL,
	path_hash BINARY(32) GENERATED ALWAYS AS (UNHEX(SHA2(path, 256))) VIRTUAL,
	faculty VARCHAR(30) NOT NULL,
    programme VARCHAR(30) NOT NULL,
	claimedBy VARCHAR(10),
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	modified DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY ux_path_hash (path_hash)

)`

const createMySQLRulesTableTmpl = `CREATE TABLE IF NOT EXISTS %s (
	id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
	directoryID INT UNSIGNED NOT NULL,
	backupType TINYTEXT NOT NULL,
	backupMetadata TEXT,
	backupFrequency SMALLINT UNSIGNED NOT NULL,
	wildcardMatch TEXT NOT NULL,
	reviewAt DATE NOT NULL,
	deleteAt DATE NOT NULL,
	created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	modified DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	
	FOREIGN KEY (directoryID) REFERENCES %s(id) ON DELETE CASCADE
)`

const mySQLShowTablesStmt = "SHOW TABLES"
const mySQLDuplicateEntryError = "Duplicate entry"

var ErrMissingArgument = errors.New("missing required argument")

type MySQLSource struct {
	*SQLSource
}

func NewMySQLSourceFromEnv(directoriesTableName, rulesTableName string,
) (*MySQLSource, error) {
	return NewMySQLSource(
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_DATABASE"),
		directoriesTableName,
		rulesTableName,
	)
}

// NewMySQLSource opens a connection to a MySQL database using given credentials and stores it internally.
// It also creates a table with the given name if it does not exist.
// You are responsible to close the connection using Close().
func NewMySQLSource(host, port, user, password, dbName, directoriesTableName,
	rulesTableName string) (*MySQLSource, error) {
	var missing []string

	appendIfEmpty(&missing, "host", host)
	appendIfEmpty(&missing, "port", port)
	appendIfEmpty(&missing, "user", user)
	appendIfEmpty(&missing, "password", password)
	appendIfEmpty(&missing, "dbName", dbName)

	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: %v\n", ErrMissingArgument, missing)
	}

	address := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName)

	db, err := sql.Open("mysql", address)
	if err != nil {
		return nil, err
	}

	sq := &MySQLSource{
		&SQLSource{
			db:                   db,
			directoriesTableName: directoriesTableName,
			rulesTableName:       rulesTableName,
			dupRowsError:         mySQLDuplicateEntryError,
		},
	}

	return sq, sq.Init()
}

func appendIfEmpty(array *[]string, name, val string) {
	if val == "" {
		*array = append(*array, name)
	}
}

func (sq MySQLSource) ShowTables() ([]string, error) {
	return sq.showTables(mySQLShowTablesStmt)
}

func (sq MySQLSource) Init() error {
	tables, err := sq.ShowTables()
	if err != nil {
		return err
	}

	initialised := true
	requiredTables := []string{sq.directoriesTableName, sq.rulesTableName}

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
		createMySQLDirectoriesTableTmpl,
		createMySQLRulesTableTmpl,
	)
}
