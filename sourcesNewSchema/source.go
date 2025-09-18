package sourcesNewSchema

import (
	"fmt"
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
	BackupType      string
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

	if r.BackupFrequency == 0 {
		return fmt.Errorf("%w: BackupFrequency", ErrMissingArgument)
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
		r.BackupFrequency = DefaultBackupFrequency
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
