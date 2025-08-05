package sources

import (
	"errors"
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
	ReportingName string      `csv:"reporting_name"`
	ReportingRoot string      `csv:"reporting_root"`
	Directory     string      `csv:"directory"`
	Instruction   Instruction `csv:"instruction"`
	Match         string      `csv:"match"`
	Ignore        string      `csv:"ignore"`
	Requestor     string      `csv:"requestor"`
	Faculty       string      `csv:"faculty"`
	ID            uint16      `csv:"id"`
}

var ErrNoEntry = errors.New("entry does not exist")
