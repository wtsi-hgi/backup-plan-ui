package converter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wtsi-hgi/backup-plan-ui/sources"
	"github.com/wtsi-hgi/backup-plan-ui/sourcesNewSchema" //nolint:goimports
)

var ErrInvalidInstruction = errors.New("invalid instruction")

func ConvertSchema(cfgSource sources.MySQLConfig, sourceTableName string,
	cfgTarget sourcesnewschema.MySQLConfig, dirsTableName, rulesTableName string, dropTable bool) error {
	oldDB, err := sources.NewMySQLSource(cfgSource, sourceTableName)
	if err != nil {
		return err
	}

	defer callAndLogError(oldDB.Close)

	newDB, err := sourcesnewschema.NewMySQLSource(cfgTarget, dirsTableName, rulesTableName)
	if err != nil {
		return err
	}

	defer callAndLogError(newDB.Close)

	if dropTable {
		err = resetDB(newDB)
		if err != nil {
			return err
		}
	}

	return convertDBSchema(oldDB, newDB)
}

func callAndLogError(f func() error) {
	err := f()
	if err != nil {
		slog.Error(err.Error())
	}
}

func resetDB(db *sourcesnewschema.MySQLSource) error {
	ctx := context.TODO()

	err := db.DropTables(ctx)
	if err != nil {
		return err
	}

	return db.Init(ctx)
}

func convertDBSchema(oldDB sources.MySQLSource, newDB *sourcesnewschema.MySQLSource) error {
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

func convertEntry(entry *sources.Entry) (sourcesnewschema.Directory, error) {
	directory := sourcesnewschema.Directory{
		Path:      entry.Directory,
		Faculty:   entry.Faculty,
		Programme: "unknown",
		ClaimedBy: entry.Requestor,
	}

	backupType, err := convertInstruction(entry.Instruction)
	if err != nil {
		return directory, err
	}

	baseRule := sourcesnewschema.Rule{
		BackupType:     backupType,
		BackupMetadata: entry.Metadata,
		WildcardMatch:  entry.Match,
	}

	baseRule.SetDefaults()

	addRulesByPattern(&directory, baseRule, baseRule.WildcardMatch)

	if entry.Ignore != "" {
		baseRule.BackupType = sourcesnewschema.NoBackup
		baseRule.BackupFrequency = 0

		addRulesByPattern(&directory, baseRule, entry.Ignore)
	}

	return directory, nil
}

func convertInstruction(instruction sources.Instruction) (sourcesnewschema.Instruction, error) {
	switch instruction {
	case sources.Backup:
		return sourcesnewschema.Backup, nil
	case sources.NoBackup:
		return sourcesnewschema.NoBackup, nil
	case sources.TempBackup:
		return sourcesnewschema.TempBackup, nil
	case sources.ManualBackup:
		return sourcesnewschema.ManualBackup, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidInstruction, instruction)
	}
}

func addRulesByPattern(directory *sourcesnewschema.Directory, baseRule sourcesnewschema.Rule, patterns string) {
	for pattern := range strings.SplitSeq(patterns, " ") {
		baseRule.WildcardMatch = pattern

		directory.AddRule(baseRule)
	}
}

func addOrUpdateDirectory(directory sourcesnewschema.Directory, db *sourcesnewschema.MySQLSource) error {
	ctx := context.TODO()

	_, err := db.AddDirectory(ctx, directory)
	if err == nil {
		return nil
	}

	if !errors.Is(err, sourcesnewschema.ErrDirectoryDuplicate) {
		return err
	}

	return updateDirectory(ctx, db, directory)
}

func updateDirectory(ctx context.Context, db *sourcesnewschema.MySQLSource, dir sourcesnewschema.Directory) error {
	existingDir, err := db.SearchDirectory(ctx, dir.Path)
	if err != nil {
		return err
	}

	if existingDir.ClaimedBy != dir.ClaimedBy {
		slog.Warn(fmt.Sprintf(
			"Directory %s was already claimed by %s, %s will lose ownership",
			dir.Path, existingDir.ClaimedBy, dir.ClaimedBy,
		))
	}

	for _, rule := range dir.Rules {
		_, err = db.SetRule(ctx, existingDir.ID, rule)
		if err != nil {
			return err
		}
	}

	return nil
}
