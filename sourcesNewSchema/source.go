package sourcesNewSchema

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultBackupFrequency = 7
	DefaultReviewMonth     = 6
	DefaultDeleteMonth     = 12
)

type DataSource interface {
	Init() error
	Close() error

	AddDirectory(directory Directory) (uint, error)
	GetDirectory(id uint) (*Directory, error)
	SearchDirectory(path string) (*Directory, error)
	ClaimDirectory(id uint, user string) error
	DeleteDirectory(id uint) error

	SetRule(id uint, rule Rule) (uint, error)
	UpdateRule(id uint, rule Rule) error
	DeleteRule(id uint) error
}

type Instruction string

const (
	Backup       Instruction = "backup"
	NoBackup     Instruction = "nobackup"
	TempBackup   Instruction = "tempbackup"
	ManualBackup Instruction = "manual backup"
)

var instructionLookup = map[string]Instruction{
	string(Backup):       Backup,
	string(NoBackup):     NoBackup,
	string(TempBackup):   TempBackup,
	string(ManualBackup): ManualBackup,
}

var ErrWrongInstruction = errors.New("wrong instruction")

// ParseInstruction parses a string into a valid Instruction.
func ParseInstruction(s string) (Instruction, error) {
	normalised := strings.TrimSpace(s)
	if v, ok := instructionLookup[normalised]; ok {
		return v, nil
	}

	return "", fmt.Errorf("%w: %s", ErrWrongInstruction, s)
}

type Directory struct {
	ID        uint
	Path      string
	Faculty   string
	Programme string
	ClaimedBy string
	Rules     []Rule
}

type Rule struct {
	ID              uint
	BackupType      Instruction
	BackupMetadata  string
	BackupFrequency int
	ReviewAt        time.Time
	DeleteAt        time.Time
	WildcardMatch   string
}

func (d *Directory) AddRule(rule Rule) {
	d.Rules = append(d.Rules, rule)
}

func (r *Rule) IsValid() error {
	if r.BackupType == "" {
		return fmt.Errorf("%w: BackupType", ErrMissingArgument)
	}

	if r.ReviewAt.IsZero() {
		return fmt.Errorf("%w: ReviewAt", ErrMissingArgument)
	}

	if r.DeleteAt.IsZero() {
		return fmt.Errorf("%w: DeleteAt", ErrMissingArgument)
	}

	if r.WildcardMatch == "" {
		return fmt.Errorf("%w: WildcardMatch", ErrMissingArgument)
	}

	return nil
}

func (r *Rule) SetDefaults() {
	if r.BackupFrequency == 0 {
		if r.BackupType != NoBackup && r.BackupType != ManualBackup {
			r.BackupFrequency = DefaultBackupFrequency
		}
	}

	now := time.Now().UTC()

	if r.ReviewAt.IsZero() {
		r.ReviewAt = now.AddDate(0, DefaultReviewMonth, 0)
	}

	if r.DeleteAt.IsZero() {
		r.DeleteAt = now.AddDate(0, DefaultDeleteMonth, 0)
	}

	if r.WildcardMatch == "" {
		r.WildcardMatch = "*"
	}
}
