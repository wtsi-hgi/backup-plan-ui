package main

import (
	"fmt"
	"html/template"
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
	ID            uint16
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	if len(os.Args) != 2 {
		fmt.Println("You should provide a path to a database")
		os.Exit(1)
	}

	dbPath := os.Args[1]

	var db DataSource

	db = CSVSource{dbPath}
	server := server{db: db}

	r := chi.NewRouter()

	r.Get("/", serveHome)
	r.Get("/hello", sayHello)
	r.Get("/models", models)

	r.Get("/entries", server.getEntries)
	r.Get("/actions/edit/{id}", server.allowUserToEditRow)
	r.Put("/actions/submit/{id}", server.submitEdits)
	r.Get("/actions/cancel/{id}", server.cancelEdit)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.ListenAndServe(":4000", r)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<p>Hello from Go! 👋</p>"))
}

func models(w http.ResponseWriter, r *http.Request) {
	make := r.URL.Query().Get("make")

	fmt.Println("Selected make:", make)

	w.Write([]byte("<option value='325i'>325i</option>\n<option value='325ix'>325ix</option>\n<option value='X5'>X5</option> "))
}
