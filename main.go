package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

//go:embed static
var staticFiles embed.FS

//go:embed templates
var templateFiles embed.FS

type instruction string

const (
	Backup     instruction = "backup"
	NoBackup   instruction = "nobackup"
	TempBackup instruction = "tempbackup"
)

type Entry struct {
	ReportingName string      `csv:"reporting_name"`
	ReportingRoot string      `csv:"reporting_root"`
	Directory     string      `csv:"directory"`
	Instruction   instruction `csv:"instruction"`
	Match         string    `csv:"match"`
	Ignore        string    `csv:"ignore"`
	Requestor     string      `csv:"requestor"`
	Faculty       string      `csv:"faculty"`
	ID            uint16      `csv:"id"`
}

var tmpl = template.Must(template.ParseFS(templateFiles, templateDir+"index.html"))

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <path-to-csv>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	dbPath := os.Args[1]
	fmt.Println("Using database:", dbPath)

	server := server{
		db: CSVSource{dbPath},
	}

	r := chi.NewRouter()

	r.Get("/", serveHome)

	r.Get("/entries", server.getEntries)
	r.Get("/actions/edit/{id}", server.allowUserToEditRow)
	r.Put("/actions/submit/{id}", server.submitEdits)
	r.Get("/actions/cancel/{id}", server.resetView)
	r.Get("/actions/delete/{id}", server.deleteRow)
	r.Get("/actions/startDelete/{id}", server.openDeleteDialog)
	r.Get("/actions/cancelDel", returnEmpty)
	r.Get("/actions/add", server.showAddRowForm)
	r.Put("/actions/add", server.addNewEntry)

	r.Handle("/static/*", http.FileServerFS(staticFiles))

	if err := http.ListenAndServe(":4000", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

var returnEmpty = func(w http.ResponseWriter, r *http.Request) {}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
	}
}
