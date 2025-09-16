package sourcesNewSchema

import (
	"database/sql"
)

const createSQLiteUsersTableTmpl = `CREATE TABLE %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    userName TEXT NOT NULL,
    faculty TEXT NOT NULL,
    programme TEXT NOT NULL
)`

const createSQLiteDirectoriesTableTmpl = `CREATE TABLE %s (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT NOT NULL,
	claimedByUserID INTEGER,
	
	FOREIGN KEY (claimedByUserID) REFERENCES %s(id) ON DELETE SET NULL
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

// NewSQLiteSource opens a connection to an SQLite database at the given path and stores it internally.
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
			usersTableName:       DefaultUsersTableName,
			directoriesTableName: DefaultDirectoriesTableName,
			rulesTableName:       DefaultRulesTableName,
		},
	}

	return sq, sq.Init()
}

func (sq SQLiteSource) ShowTables() ([]string, error) {
	return sq.showTables(showSQLiteTablesStmt)
}

func (sq SQLiteSource) Init() error {
	return sq.init(
		createSQLiteUsersTableTmpl,
		createSQLiteDirectoriesTableTmpl,
		createSQLiteRulesTableTmpl,
	)
}
