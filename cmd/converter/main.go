package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wtsi-hgi/backup-plan-ui/converter"
	"github.com/wtsi-hgi/backup-plan-ui/sources"
)

func usage() {
	prog := filepath.Base(os.Args[0])
	fmt.Println("Add data from CSV to SQLite or MySQL database.")
	fmt.Println("\nUsage:")
	fmt.Printf("  %s -b sqlite --csv <path-to-csv> --sqlite <path-to-sqlite> [--replace]\n", prog)
	fmt.Printf("  %s -b mysql --csv <path-to-csv> [--table table-name] [--replace]\n", prog)
	fmt.Println("\nEnvironment (mysql): MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASS, MYSQL_DATABASE")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
}

var (
	backend    string
	csvPath    string
	sqlitePath string
	tableName  string
	dropTable  bool
)

func init() {
	flag.StringVar(&backend, "b", "", "Backend to use (sqlite or mysql)")
	flag.StringVar(&csvPath, "csv", "", "Path to CSV file")
	flag.StringVar(&sqlitePath, "sqlite", "", "Path to SQLite file")
	flag.BoolVar(&dropTable, "replace", false, "Remove existing data before inserting new data")
	flag.StringVar(&tableName, "table", sources.DefaultTableName, "Name of table to insert data into")

	flag.Usage = usage
}

func main() {
	flag.Parse()

	if csvPath == "" {
		log.Fatalf("You must specify a CSV file.")
	}

	switch backend {
	case "sqlite":
		if sqlitePath == "" {
			log.Fatalf("You must specify a SQLite file.")
		}

		if err := converter.ConvertCsvToSqlite(csvPath, sqlitePath, dropTable); err != nil {
			log.Fatalf("Conversion failed: %v", err)
		}

	case "mysql":
		if err := converter.ConvertCsvToMySQL(csvPath, tableName, dropTable); err != nil {
			log.Fatalf("Conversion failed: %v", err)
		}

	default:
		log.Fatalf("Invalid backend: %s", backend)
	}

	fmt.Println("Data conversion was successful.")
}
