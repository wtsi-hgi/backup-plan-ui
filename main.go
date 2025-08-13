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

	"github.com/go-chi/chi/v5"
)

//go:embed static
var staticFiles embed.FS

//go:embed templates
var templateFiles embed.FS

func main() {
	log.SetFlags(0) // timestamp comes from systemd

	db := parseArgs(os.Args[1:])

	srv, err := server.NewServer(db, templateFiles)
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

func parseArgs(args []string) sources.DataSource {
	if len(args) == 0 {
		usage("Not enough arguments.")
	}

	backend := args[0]

	var (
		db  sources.DataSource
		msg string
		err error
	)

	switch {
	case backend == "mysql":
		db, err = sources.NewMySQLSource(
			os.Getenv("MYSQL_HOST"),
			os.Getenv("MYSQL_PORT"),
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASS"),
			os.Getenv("MYSQL_DATABASE"),
			sources.DefaultTableName,
		)
		msg = "Using MySQL database"
	case len(args) == 1:
		usage("Not enough arguments.")
	case backend == "sqlite":
		db, err = sources.NewSQLiteSource(args[1])
		msg = "Using SQLite database: " + args[1]
	case backend == "csv":
		db = sources.CSVSource{Path: args[1]}
		msg = "Using CSV file: " + args[1]
	default:
		usage("Arguments are not recognized.")
	}

	if err != nil {
		log.Fatal(err)
	}

	slog.Info(msg)

	return db
}

func usage(msg string) {
	if msg != "" {
		slog.Error(msg)
	}
	fmt.Println("Usage:")
	fmt.Println("  backup-plan-ui csv <path/to/file.csv>")
	fmt.Println("  backup-plan-ui sqlite <path/to/file.sqlite>")
	fmt.Println("  backup-plan-ui mysql")
	os.Exit(2)
}

var returnEmpty = func(w http.ResponseWriter, r *http.Request) {}
