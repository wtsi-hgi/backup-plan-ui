package converter

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/wtsi-hgi/backup-plan-ui/sources"
	"github.com/wtsi-hgi/backup-plan-ui/sourcesNewSchema"
)

func ConvertSchema(sourceTableName, dirsTableName, rulesTableName string, dropTable bool) error {
	oldDB, err := sources.NewMySQLSourceFromEnv(sourceTableName)
	if err != nil {
		return err
	}

	defer callAndLogError(oldDB.Close)

	newDB, err := sourcesNewSchema.NewMySQLSourceFromEnv(dirsTableName, rulesTableName)
	if err != nil {
		return err
	}

	defer callAndLogError(newDB.Close)

	return convertDBSchema(oldDB, newDB)
}

func callAndLogError(f func() error) {
	err := f()
	if err != nil {
		slog.Error(err.Error())
	}
}

func convertDBSchema(oldDB sources.MySQLSource, newDB *sourcesNewSchema.MySQLSource) error {
	entries, err := oldDB.ReadAll()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		directory, err := convertEntry(entry)
		if err != nil {
			return err
		}

		err = addOrUpdateDirectory(directory, newDB)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertEntry(entry *sources.Entry) (sourcesNewSchema.Directory, error) {
	directory := sourcesNewSchema.Directory{
		Path:      entry.Directory,
		Faculty:   entry.Faculty,
		Programme: "unknown",
		ClaimedBy: entry.Requestor,
	}

	rule := sourcesNewSchema.Rule{
		BackupType:     string(entry.Instruction),
		BackupMetadata: entry.Metadata,
		WildcardMatch:  entry.Match,
	}

	rule.SetDefaults()

	directory.AddRule(rule)

	if entry.Ignore != "" {
		rule.BackupType = string(sources.NoBackup)
		rule.WildcardMatch = entry.Ignore

		directory.AddRule(rule)
	}

	return directory, nil
}

func addOrUpdateDirectory(directory sourcesNewSchema.Directory, db *sourcesNewSchema.MySQLSource) error {
	_, err := db.AddDirectory(directory)
	if err == nil {
		return nil
	}

	if !errors.Is(err, sourcesNewSchema.ErrDirectoryDuplicate) {
		return err
	}

	existingDir, err := db.SearchDirectory(directory.Path)
	if err != nil {
		return err
	}

	if existingDir.ClaimedBy != directory.ClaimedBy {
		slog.Warn(fmt.Sprintf(
			"Directory %s was already claimed by %s, %s will lose ownership",
			directory.Path, existingDir.ClaimedBy, directory.ClaimedBy,
		))
	}

	for _, rule := range directory.Rules {
		_, err = db.SetRule(existingDir.ID, rule)
		if err != nil {
			return err
		}
	}

	return nil
}
