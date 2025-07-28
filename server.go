package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type server struct {
	db DataSource
}

var (
	tmplRow     = parseTemplate("row.html")
	tmplEditRow = parseTemplate("edit_row.html")
)

const templateDir = "templates/"

func parseTemplate(name string) *template.Template {
	return template.Must(
		template.New(name).
			Funcs(template.FuncMap{
				"join": joinCommaSpace,
			}).
			ParseFiles(templateDir + name))
}

func joinCommaSpace(items []string) string {
	return strings.Join(items, ", ")
}

func (s server) getEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := s.db.readAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	for _, entry := range entries {
		tmplRow.Execute(w, entry)
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

	return tmpl.Execute(w, entry)
}

func (s server) resetView(w http.ResponseWriter, r *http.Request) {
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

	r.ParseForm()

	updatedEntry := createEntryFromForm(uint16(id), r)

	err = s.db.updateEntry(updatedEntry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	s.resetView(w, r)
}

func createEntryFromForm(id uint16, r *http.Request) *Entry {
	return &Entry{
		ID:            id,
		ReportingName: r.FormValue("ReportingName"),
		ReportingRoot: r.FormValue("ReportingRoot"),
		Directory:     r.FormValue("Directory"),
		Instruction:   instruction(r.FormValue("Instruction")),
		Match:         splitList(r.FormValue("Match")),
		Ignore:        splitList(r.FormValue("Ignore")),
		Requestor:     r.FormValue("Requestor"),
		Faculty:       r.FormValue("Faculty"),
	}
}

func splitList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n'
	})

	var cleaned []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}

	return cleaned
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
}
