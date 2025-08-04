package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type server struct {
	db DataSource
}

var (
	tmplRow          = parseTemplate("templates/row.html")
	tmplEditRow      = parseTemplate("templates/edit_row.html")
	tmplAddRow       = parseTemplate("templates/add_row.html")
	tmplDeleteDialog = parseTemplate("templates/delete_modal.html")
)

type formField string

const (
	ErrBlankInput                 = "You cannot leave this field blank"
	ErrInvalidInstruction         = "Input must be backup, tempBackup or noBackup"
	ErrIgnoreWithoutBackup        = "Ignore can only be used with the backup instruction"
	ErrDirectoryNotInRoot         = "Directory must be inside Reporting root"
	ErrReportingRootNotDeepEnough = "Reporting Root must be atleast five levels deep"
	ErrRootWithoutSlash           = "Reporting Root must start with a slash (/)"

	ReportingName formField = "ReportingName"
	ReportingRoot formField = "ReportingRoot"
	Directory     formField = "Directory"
	Instruction   formField = "Instruction"
	Match         formField = "Match"
	Ignore        formField = "Ignore"
	Requestor     formField = "Requestor"
	Faculty       formField = "Faculty"
)

func (f formField) string() string {
	return string(f)
}

func parseTemplate(name string) *template.Template {
	return template.Must(template.ParseFS(templateFiles, name))
}

type tmplData struct {
	Entry  *Entry
	Errors map[formField]string
}

func (s server) getEntries(w http.ResponseWriter, _ *http.Request) {
	entries, err := s.db.readAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	for _, entry := range entries {
		err = tmplRow.Execute(w, tmplData{Entry: entry})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s server) allowUserToEditRow(w http.ResponseWriter, r *http.Request) {
	err := s.changeTemplate(w, r, tmplEditRow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s server) changeTemplate(w http.ResponseWriter, r *http.Request, tmpl *template.Template) error {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	entry, err := s.db.getEntry(uint16(id))
	if err != nil {
		return err
	}

	return tmpl.Execute(w, tmplData{Entry: entry})
}

func (s server) resetView(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	if idStr == "new" {
		return
	}

	err := s.changeTemplate(w, r, tmplRow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s server) submitEdits(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	validationErrors := validateForm(r)
	updatedEntry := createEntryFromForm(uint16(id), r)

	if len(validationErrors) > 0 {
		data := tmplData{
			Entry:  updatedEntry,
			Errors: validationErrors,
		}

		err := tmplEditRow.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	err = s.db.updateEntry(updatedEntry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	s.resetView(w, r)
}

func validateForm(r *http.Request) map[formField]string {
	errors := make(map[formField]string)

	validateNonBlankInputs(r, errors)
	validateInstructionAndIgnore(r, errors)
	validateDirectoryAndRoot(r, errors)

	return errors
}

func validateNonBlankInputs(r *http.Request, errors map[formField]string) {
	requiredFields := []formField{ReportingName, ReportingRoot, Directory,
		Instruction, Requestor, Faculty}

	for _, requiredField := range requiredFields {
		if r.FormValue(requiredField.string()) == "" {
			errors[requiredField] = ErrBlankInput
		}
	}
}

func validateInstructionAndIgnore(r *http.Request, errors map[formField]string) {
	instr := instruction(r.FormValue(Instruction.string()))
	ignore := r.FormValue(Ignore.string())

	if instr != Backup && instr != TempBackup && instr != NoBackup {
		addToMapIfNew(errors, Instruction, ErrInvalidInstruction)
	}

	if ignore != "" && instr != Backup {
		addToMapIfNew(errors, Ignore, ErrIgnoreWithoutBackup)
	}
}

func addToMapIfNew(givenMap map[formField]string, key formField, value string) {
	if _, exists := givenMap[key]; !exists {
		givenMap[key] = value
	}
}

func validateDirectoryAndRoot(r *http.Request, errors map[formField]string) {
	reportingRoot := r.FormValue(ReportingRoot.string())
	dir := r.FormValue(Directory.string())

	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	if !strings.HasPrefix(reportingRoot, "/") {
		addToMapIfNew(errors, ReportingRoot, ErrRootWithoutSlash)
	}

	rel, err := filepath.Rel(reportingRoot, dir)
	if err != nil || strings.HasPrefix(rel, "../") || rel == ".." {
		addToMapIfNew(errors, Directory, ErrDirectoryNotInRoot)
	}

	depth := 0
	for _, part := range strings.Split(reportingRoot, string(filepath.Separator)) {
		if part != "" {
			depth++
		}
	}

	if depth < 5 {
		addToMapIfNew(errors, ReportingRoot, ErrReportingRootNotDeepEnough)
	}
}

func createEntryFromForm(id uint16, r *http.Request) *Entry {
	return &Entry{
		ID:            id,
		ReportingName: r.FormValue(ReportingName.string()),
		ReportingRoot: r.FormValue(ReportingRoot.string()),
		Directory:     r.FormValue(Directory.string()),
		Instruction:   instruction(r.FormValue(Instruction.string())),
		Match:         r.FormValue(Match.string()),
		Ignore:        r.FormValue(Ignore.string()),
		Requestor:     r.FormValue(Requestor.string()),
		Faculty:       r.FormValue(Faculty.string()),
	}
}

func (s server) deleteRow(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = s.db.deleteEntry(uint16(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(fmt.Sprintf(`
		<script>
			document.getElementById('modal')?.remove();
			document.querySelector('tr[data-id="%d"]')?.remove();
		</script>
	`, id)))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s server) showAddRowForm(w http.ResponseWriter, _ *http.Request) {
	err := tmplAddRow.Execute(w, tmplData{Entry: &Entry{}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s server) addNewEntry(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	validationErrors := validateForm(r)

	var dummyEntryID uint16 // will be set later
	newEntry := createEntryFromForm(dummyEntryID, r)

	if len(validationErrors) > 0 {
		data := tmplData{
			Entry:  newEntry,
			Errors: validationErrors,
		}

		err := tmplAddRow.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = s.db.addEntry(newEntry)
	if err != nil {
		http.Error(w, "Failed to add entry: "+err.Error(), http.StatusInternalServerError)

		return
	}

	// Set HX-Trigger to refresh the entry table
	w.Header().Set("HX-Trigger", "entriesChanged")
}

func (s server) openDeleteDialog(w http.ResponseWriter, r *http.Request) {
	err := s.changeTemplate(w, r, tmplDeleteDialog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
