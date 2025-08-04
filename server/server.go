package server

import (
	"backup-plan-ui/sources"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	db        sources.DataSource
	templates *template.Template
}

const templatesDir = "templates"

func NewServer(db sources.DataSource, fs embed.FS) (*Server, error) {
	t, err := template.ParseFS(fs, filepath.Join(templatesDir, "*.html"))

	return &Server{
		db:        db,
		templates: t,
	}, err
}

type formField string

func (f formField) string() string {
	return string(f)
}

const (
	ReportingName formField = "ReportingName"
	ReportingRoot formField = "ReportingRoot"
	Directory     formField = "Directory"
	Instruction   formField = "Instruction"
	Match         formField = "Match"
	Ignore        formField = "Ignore"
	Requestor     formField = "Requestor"
	Faculty       formField = "Faculty"

	tmplRowPath          = "row.html"
	tmplEditRowPath      = "edit_row.html"
	tmplAddRowPath       = "add_row.html"
	tmplDeleteDialogPath = "delete_modal.html"
	tmplIndexPath        = "index.html"
)

type tmplData struct {
	Entry  *sources.Entry
	Errors map[string]string
}

func (s Server) ServeHome(w http.ResponseWriter, _ *http.Request) {
	if err := s.templates.ExecuteTemplate(w, tmplIndexPath, nil); err != nil {
		s.abortWithError(w, err, http.StatusInternalServerError)
	}
}

func (s Server) abortWithError(w http.ResponseWriter, err error, statusCode int) {
	slog.Error(err.Error())
	http.Error(w, err.Error(), statusCode)
}

func (s Server) GetEntries(w http.ResponseWriter, _ *http.Request) {
	entries, err := s.db.ReadAll()
	if err != nil {
		s.abortWithError(w, err, http.StatusInternalServerError)

		return
	}

	for _, entry := range entries {
		err = s.templates.ExecuteTemplate(w, tmplRowPath, tmplData{Entry: entry})
		if err != nil {
			s.abortWithError(w, err, http.StatusInternalServerError)
		}
	}
}

func (s Server) AllowUserToEditRow(w http.ResponseWriter, r *http.Request) {
	err := s.changeTemplate(w, r, tmplEditRowPath)
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)
	}
}

func (s Server) changeTemplate(w http.ResponseWriter, r *http.Request, tmplPath string) error {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	entry, err := s.db.GetEntry(uint16(id))
	if err != nil {
		return err
	}

	return s.templates.ExecuteTemplate(w, tmplPath, tmplData{Entry: entry})
}

func (s Server) ResetView(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	if idStr == "new" {
		return
	}

	err := s.changeTemplate(w, r, tmplRowPath)
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)
	}
}

func (s Server) SubmitEdits(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)

		return
	}

	err = r.ParseForm()
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)

		return
	}

	validationErrors := validateForm(r)
	updatedEntry := createEntryFromForm(uint16(id), r)

	if len(validationErrors) > 0 {
		data := tmplData{
			Entry:  updatedEntry,
			Errors: convertErrors(validationErrors),
		}

		err := s.templates.ExecuteTemplate(w, tmplEditRowPath, data)
		if err != nil {
			s.abortWithError(w, err, http.StatusInternalServerError)
		}

		return
	}

	err = s.db.UpdateEntry(updatedEntry)
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)

		return
	}

	slog.Info(fmt.Sprintf("Updated entry: %+v\n", *updatedEntry))

	s.ResetView(w, r)
}

func createEntryFromForm(id uint16, r *http.Request) *sources.Entry {
	return &sources.Entry{
		ID:            id,
		ReportingName: r.FormValue(ReportingName.string()),
		ReportingRoot: r.FormValue(ReportingRoot.string()),
		Directory:     r.FormValue(Directory.string()),
		Instruction:   sources.Instruction(r.FormValue(Instruction.string())),
		Match:         r.FormValue(Match.string()),
		Ignore:        r.FormValue(Ignore.string()),
		Requestor:     r.FormValue(Requestor.string()),
		Faculty:       r.FormValue(Faculty.string()),
	}
}

func convertErrors(errs map[formField]string) map[string]string {
	result := make(map[string]string, len(errs))

	for k, v := range errs {
		result[k.string()] = v
	}

	return result
}

func (s Server) DeleteRow(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)

		return
	}

	err = s.db.DeleteEntry(uint16(id))
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)
	}

	slog.Info(fmt.Sprintf("Deleted entry with id %d\n", id))

	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(fmt.Sprintf(`
		<script>
			document.getElementById('modal')?.remove();
			document.querySelector('tr[data-id="%d"]')?.remove();
		</script>
	`, id)))

	if err != nil {
		s.abortWithError(w, err, http.StatusInternalServerError)
	}
}

func (s Server) ShowAddRowForm(w http.ResponseWriter, _ *http.Request) {
	err := s.templates.ExecuteTemplate(w, tmplAddRowPath, tmplData{Entry: &sources.Entry{}})
	if err != nil {
		s.abortWithError(w, err, http.StatusInternalServerError)
	}
}

func (s Server) AddNewEntry(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)

		return
	}

	validationErrors := validateForm(r)

	var dummyEntryID uint16 // will be set later
	newEntry := createEntryFromForm(dummyEntryID, r)

	if len(validationErrors) > 0 {
		data := tmplData{
			Entry:  newEntry,
			Errors: convertErrors(validationErrors),
		}

		err := s.templates.ExecuteTemplate(w, tmplAddRowPath, data)
		if err != nil {
			s.abortWithError(w, err, http.StatusInternalServerError)
		}
		return
	}

	err = s.db.AddEntry(newEntry)
	if err != nil {
		s.abortWithError(w, err, http.StatusInternalServerError)

		return
	}

	slog.Info(fmt.Sprintf("Added new entry: %+v\n", *newEntry))

	// Set HX-Trigger to refresh the entry table
	w.Header().Set("HX-Trigger", "entriesChanged")
}

func (s Server) OpenDeleteDialog(w http.ResponseWriter, r *http.Request) {
	err := s.changeTemplate(w, r, tmplDeleteDialogPath)
	if err != nil {
		s.abortWithError(w, err, http.StatusBadRequest)
	}
}
