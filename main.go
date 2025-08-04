package main

import (
	"backup-plan-ui/server"
	"backup-plan-ui/sources"
	"embed"
	"fmt"
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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <path-to-csv>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	dbPath := os.Args[1]
	fmt.Println("Using database:", dbPath)

	server, err := server.NewServer(sources.CSVSource{Path: dbPath}, templateFiles)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("BACKUP_PLAN_UI_PORT")
	if port == "" {
		port = "4000"
	}

	r := chi.NewRouter()

	r.Get("/", server.ServeHome)

	r.Get("/entries", server.GetEntries)
	r.Get("/actions/edit/{id}", server.AllowUserToEditRow)
	r.Put("/actions/submit/{id}", server.SubmitEdits)
	r.Get("/actions/cancel/{id}", server.ResetView)
	r.Get("/actions/delete/{id}", server.DeleteRow)
	r.Get("/actions/startDelete/{id}", server.OpenDeleteDialog)
	r.Get("/actions/cancelDel", returnEmpty)
	r.Get("/actions/add", server.ShowAddRowForm)
	r.Put("/actions/add", server.AddNewEntry)

	r.Handle("/static/*", http.FileServerFS(staticFiles))

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

var returnEmpty = func(w http.ResponseWriter, r *http.Request) {}
