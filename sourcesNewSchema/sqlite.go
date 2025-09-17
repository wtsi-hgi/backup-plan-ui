package sourcesNewSchema

import (
	"database/sql"
)

const createSQLiteDirectoriesTableTmpl = `CREATE TABLE %s (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT NOT NULL UNIQUE,
	faculty TEXT NOT NULL,
    programme TEXT NOT NULL,
	claimedBy TEXT
)`

const createSQLiteRulesTableTmpl = `CREATE TABLE %s (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	directoryID INTEGER NOT NULL,
	backupType TEXT NOT NULL,
	backupMetadata TEXT,
	backupFrequency INTEGER NOT NULL,
	reviewAt INTEGER,
	deleteAt INTEGER,
	wildcardMatch TEXT,
	
	FOREIGN KEY (directoryID) REFERENCES %s(id) ON DELETE CASCADE
)`

const showSQLiteTablesStmt = "SELECT name FROM sqlite_master WHERE type='table'"

type SQLiteSource struct {
	*SQLSource
}

// NewSQLiteSource opens a connection to an SQLite database at the given Path and stores it internally.
// It also creates a table with the given name if it does not exist.
// You are responsible to close the connection using Close().
func NewSQLiteSource(path string) (*SQLiteSource, error) {
	// TODO add test for foreign keys
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	sq := &SQLiteSource{
		&SQLSource{
			db:                   db,
			directoriesTableName: DefaultDirectoriesTableName,
			rulesTableName:       DefaultRulesTableName,
			dupRowsError:         "UNIQUE constraint failed",
		},
	}

	return sq, sq.Init()
}

func (sq SQLiteSource) ShowTables() ([]string, error) {
	return sq.showTables(showSQLiteTablesStmt)
}

func (sq SQLiteSource) Init() error {
	return sq.init(
		createSQLiteDirectoriesTableTmpl,
		createSQLiteRulesTableTmpl,
	)
}
