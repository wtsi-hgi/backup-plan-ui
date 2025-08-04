package main

import (
	"backup-plan-ui/server"
	"backup-plan-ui/sources"
	"embed"
	"fmt"
	"log"
	"log/slog"
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

	log.SetFlags(log.Lshortfile) // timestamp comes from systemd

	dbPath := os.Args[1]
	slog.Info("Using database: " + dbPath)

	srv, err := server.NewServer(sources.CSVSource{Path: dbPath}, templateFiles)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("BACKUP_PLAN_UI_PORT")
	if port == "" {
		port = "4000"
	}

	r := chi.NewRouter()

	r.Get("/", srv.ServeHome)

	r.Get("/entries", srv.GetEntries)
	r.Get("/actions/edit/{id}", srv.AllowUserToEditRow)
	r.Put("/actions/submit/{id}", srv.SubmitEdits)
	r.Get("/actions/cancel/{id}", srv.ResetView)
	r.Get("/actions/delete/{id}", srv.DeleteRow)
	r.Get("/actions/startDelete/{id}", srv.OpenDeleteDialog)
	r.Get("/actions/cancelDel", returnEmpty)
	r.Get("/actions/add", srv.ShowAddRowForm)
	r.Put("/actions/add", srv.AddNewEntry)

	r.Handle("/static/*", http.FileServerFS(staticFiles))

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

var returnEmpty = func(w http.ResponseWriter, r *http.Request) {}
