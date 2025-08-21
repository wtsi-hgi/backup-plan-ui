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
	fmt.Printf("  %s sqlite --csv <path-to-csv> --sqlite <path-to-sqlite> [--drop]\n", prog)
	fmt.Printf("  %s mysql --csv <path-to-csv> [--table table-name] [--drop]\n", prog)
	fmt.Println("\nEnvironment (mysql): MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASS, MYSQL_DATABASE")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
}

var (
	csvPath    string
	sqlitePath string
	tableName  string
	dropTable  bool
)

func init() {
	flag.StringVar(&csvPath, "csv", "", "Path to CSV file")
	flag.StringVar(&sqlitePath, "sqlite", "", "Path to SQLite file")
	flag.BoolVar(&dropTable, "drop", false, "Drop table before inserting data")
	flag.StringVar(&tableName, "table", sources.DefaultTableName, "Name of table to insert data into")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	flag.Usage = usage
	flag.Parse()

	if csvPath == "" {
		usage()
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "sqlite":
		if sqlitePath == "" {
			usage()
			os.Exit(1)
		}

		if err := converter.ConvertCsvToSqlite(csvPath, sqlitePath); err != nil {
			log.Fatalf("Conversion failed: %v", err)
		}

	case "mysql":
		host := os.Getenv("MYSQL_HOST")
		port := os.Getenv("MYSQL_PORT")
		user := os.Getenv("MYSQL_USER")
		pass := os.Getenv("MYSQL_PASS")
		db := os.Getenv("MYSQL_DATABASE")

		if err := converter.ConvertCsvToMySQL(csvPath, host, port, user, pass, db, tableName); err != nil {
			log.Fatalf("Conversion failed: %v", err)
		}

	default:
		usage()
		os.Exit(1)
	}

	fmt.Println("Data conversion was successful.")
}
