package server

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wtsi-hgi/backup-plan-ui/sources"
)

type FormValidator struct {
	request *http.Request
	errors  map[formField]string
}

type formValidationErrors map[formField]string

const (
	ErrBlankInput                 = "You cannot leave this field blank"
	ErrInvalidInstruction         = "Input must be backup, tempBackup or noBackup"
	ErrIgnoreWithoutBackup        = "Ignore can only be used with the backup instruction"
	ErrDirectoryNotInRoot         = "Directory must be inside Reporting root"
	ErrReportingRootNotDeepEnough = "Reporting Root must be atleast five levels deep"
	ErrRootWithoutSlash           = "Reporting Root must start with a slash (/)"
	ErrMetadataForNonManualSet    = "Metadata is only for 'manual backup's"
	ErrInvalidMetadata            = "Metadata field must not contain comma, newline or tab characters"
)

func validateForm(r *http.Request) formValidationErrors {
	fv := FormValidator{
		request: r,
		errors:  make(formValidationErrors),
	}

	fv.validateNonBlankInputs()
	fv.validateInstructionAndIgnore()
	fv.validateDirectoryAndRoot()
	fv.validateMetadata()

	return fv.errors
}

func (fv FormValidator) validateNonBlankInputs() {
	requiredFields := []formField{ReportingName, ReportingRoot, Directory,
		Instruction, Requestor, Faculty}

	for _, requiredField := range requiredFields {
		if fv.getFormValue(requiredField) == "" {
			fv.errors[requiredField] = ErrBlankInput
		}
	}
}

func (fv FormValidator) getFormValue(field formField) string {
	return fv.request.FormValue(field.string())
}

func (fv FormValidator) validateInstructionAndIgnore() {
	instr, err := sources.ParseInstruction(fv.getFormValue(Instruction))
	if err != nil {
		fv.addErrorIfNew(Instruction, ErrInvalidInstruction)
	}

	ignore := fv.getFormValue(Ignore)
	if ignore != "" && instr != sources.Backup {
		fv.addErrorIfNew(Ignore, ErrIgnoreWithoutBackup)
	}
}

func (fv FormValidator) addErrorIfNew(field formField, err string) {
	if _, exists := fv.errors[field]; !exists {
		fv.errors[field] = err
	}
}

func (fv FormValidator) validateDirectoryAndRoot() {
	reportingRoot := fv.getFormValue(ReportingRoot)
	dir := fv.getFormValue(Directory)

	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	if !strings.HasPrefix(reportingRoot, "/") {
		fv.addErrorIfNew(ReportingRoot, ErrRootWithoutSlash)
	}

	rel, err := filepath.Rel(reportingRoot, dir)
	if err != nil || strings.HasPrefix(rel, "../") || rel == ".." {
		fv.addErrorIfNew(Directory, ErrDirectoryNotInRoot)
	}

	depth := 0
	for _, part := range strings.Split(reportingRoot, string(filepath.Separator)) {
		if part != "" {
			depth++
		}
	}

	if depth < 5 {
		fv.addErrorIfNew(ReportingRoot, ErrReportingRootNotDeepEnough)
	}
}

func (fv FormValidator) validateMetadata() {
	instruction := fv.getFormValue(Instruction)
	metadata := fv.getFormValue(Metadata)
	if instruction != string(sources.ManualBackup) && metadata != "" {
		fv.addErrorIfNew(Metadata, ErrMetadataForNonManualSet)
	} else {
		if strings.ContainsAny(metadata, ",\n\t") {
			fv.addErrorIfNew(Metadata, ErrInvalidMetadata)
		}
	}

}
