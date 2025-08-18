package sources

import (
	"errors"
	"fmt"
	"regexp"
)

type DataSource interface {
	ReadAll() ([]*Entry, error)
	GetEntry(id uint16) (*Entry, error)
	UpdateEntry(newEntry *Entry) error
	DeleteEntry(id uint16) (*Entry, error)
	AddEntry(entry *Entry) error
}

type Instruction string

const (
	Backup     Instruction = "backup"
	NoBackup   Instruction = "nobackup"
	TempBackup Instruction = "tempbackup"
)

type Entry struct {
	ID            uint16      `csv:"id"`
	ReportingName string      `csv:"reporting_name"`
	ReportingRoot string      `csv:"reporting_root"`
	Directory     string      `csv:"directory"`
	Instruction   Instruction `csv:"instruction"`
	Frequency     string      `csv:"frequency"`
	Match         string      `csv:"match"`
	Ignore        string      `csv:"ignore"`
	Requestor     string      `csv:"requestor"`
	Faculty       string      `csv:"faculty"`
}

func ParseFrequency(freq string) (string, error) {
	var value string
	re := regexp.MustCompile(`^[1-9][0-9]*[dwm]$`)

	switch freq {
	case "never":
		value = "0"
	case "daily":
		value = "1d"
	case "weekly":
		value = "1w"
	case "monthly":
		value = "1m"
	default:
		if !re.MatchString(freq) {
			return value, fmt.Errorf("%w: %s", ErrInvalidFrequency, freq)
		}

		value = freq
	}

	return value, nil
}

var ErrNoEntry = errors.New("entry does not exist")
var ErrInvalidFrequency = errors.New("invalid frequency")
