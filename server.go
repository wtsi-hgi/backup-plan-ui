package main

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type server struct {
	db DataSource
}

var (
	tmplRow     = template.Must(template.ParseFiles("templates/row.html"))
	tmplEditRow = template.Must(template.ParseFiles("templates/edit_row.html"))
)

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

func (s server) cancelEdit(w http.ResponseWriter, r *http.Request) {
	err := s.changeTemplate(w, r, tmplRow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// doesnt update it yet
func (s server) submitEdits(w http.ResponseWriter, r *http.Request) {
	// idStr := chi.URLParam(r, "id")
	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	http.Error(w, "Invalid ID", http.StatusBadRequest)
	// 	return
	// }
	// r.ParseForm()

	// email := r.FormValue("email")

	// entry, err := s.db.getEntry(uint16(id))
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	s.cancelEdit(w, r)
}
