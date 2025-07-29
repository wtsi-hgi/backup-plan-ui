package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

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
	Match         []string    `csv:"match"`
	Ignore        []string    `csv:"ignore"`
	Requestor     string      `csv:"requestor"`
	Faculty       string      `csv:"faculty"`
	ID            uint16      `csv:"id"`
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . <path-to-csv>")
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
	r.Get("/actions/add", server.showAddRowForm)
	r.Put("/actions/add", server.addNewEntry)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	if err := http.ListenAndServe(":4000", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template rendering failed", http.StatusInternalServerError)
	}
}
