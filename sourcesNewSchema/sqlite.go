package sourcesNewSchema

import (
	"database/sql"
	"fmt"
)

const createSQLiteDirectoriesTableTmpl = `CREATE TABLE %s (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT NOT NULL UNIQUE,
	faculty TEXT NOT NULL,
    programme TEXT NOT NULL,
	claimedBy TEXT,
	created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	modified DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

const createSQLiteTriggerTmpl = `CREATE TRIGGER %[1]s_modified_timestamp
AFTER UPDATE ON %[1]s
FOR EACH ROW
BEGIN
	UPDATE %[1]s SET modified = CURRENT_TIMESTAMP WHERE id = OLD.id;
END
`

const createSQLiteRulesTableTmpl = `CREATE TABLE %s (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	directoryID INTEGER NOT NULL,
	backupType TEXT NOT NULL,
	backupMetadata TEXT,
	backupFrequency INTEGER NOT NULL,
	wildcardMatch TEXT NOT NULL,
	reviewAt DATE NOT NULL,
	deleteAt DATE NOT NULL,
	created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	modified DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	
	FOREIGN KEY (directoryID) REFERENCES %s(id) ON DELETE CASCADE
)`

const sqliteShowTablesStmt = "SELECT name FROM sqlite_master WHERE type='table'"
const sqliteDuplicateEntryError = "UNIQUE constraint failed"

type SQLiteSource struct {
	*SQLSource
}

// NewSQLiteSource opens a connection to an SQLite database at the given Path and stores it internally.
// It also creates a table with the given name if it does not exist.
// You are responsible to close the connection using Close().
func NewSQLiteSource(path string) (*SQLiteSource, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	sq := &SQLiteSource{
		&SQLSource{
			db:                   db,
			directoriesTableName: DefaultDirectoriesTableName,
			rulesTableName:       DefaultRulesTableName,
			dupRowsError:         sqliteDuplicateEntryError,
		},
	}

	return sq, sq.Init()
}

func (sq SQLiteSource) ShowTables() ([]string, error) {
	return sq.showTables(sqliteShowTablesStmt)
}

func (sq SQLiteSource) Init() error {
	err := sq.init(
		createSQLiteDirectoriesTableTmpl,
		createSQLiteRulesTableTmpl,
	)
	if err != nil {
		return err
	}

	for _, tableName := range []string{sq.directoriesTableName, sq.rulesTableName} {
		_, err = sq.db.Exec(fmt.Sprintf(createSQLiteTriggerTmpl, tableName))
		if err != nil {
			return err
		}
	}

	return nil
}
