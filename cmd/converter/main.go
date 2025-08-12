package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"backup-plan-ui/converter"
	"backup-plan-ui/sources"
)

func usage() {
	prog := filepath.Base(os.Args[0])
	fmt.Println("Usage:")
	fmt.Printf("  %s sqlite <path-to-csv> <path-to-sqlite>\n", prog)
	fmt.Printf("  %s mysql <path-to-csv> [table-name]\n", prog)
	fmt.Println("\nEnvironment (mysql): MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASS, MYSQL_DATABASE")
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "sqlite":
		if len(os.Args) != 4 {
			usage()
			os.Exit(1)
		}

		csvPath := os.Args[2]
		sqlitePath := os.Args[3]
		if err := converter.ConvertCsvToSqlite(csvPath, sqlitePath); err != nil {
			log.Fatalf("Conversion failed: %v", err)
		}

	case "mysql":
		if len(os.Args) < 3 || len(os.Args) > 4 {
			usage()
			os.Exit(1)
		}

		csvPath := os.Args[2]
		tableName := sources.DefaultTableName
		if len(os.Args) == 4 && os.Args[3] != "" {
			tableName = os.Args[3]
		}

		host := os.Getenv("MYSQL_HOST")
		port := os.Getenv("MYSQL_PORT")
		user := os.Getenv("MYSQL_USER")
		pass := os.Getenv("MYSQL_PASS")
		db := os.Getenv("MYSQL_DATABASE")

		var missing []string
		if host == "" {
			missing = append(missing, "MYSQL_HOST")
		}
		if port == "" {
			missing = append(missing, "MYSQL_PORT")
		}
		if user == "" {
			missing = append(missing, "MYSQL_USER")
		}
		if pass == "" {
			missing = append(missing, "MYSQL_PASS")
		}
		if db == "" {
			missing = append(missing, "MYSQL_DATABASE")
		}
		if len(missing) > 0 {
			log.Fatalf("Missing required environment variables: %v\n", missing)
		}

		if err := converter.ConvertCsvToMySQL(csvPath, host, port, user, pass, db, tableName); err != nil {
			log.Fatalf("Conversion failed: %v", err)
		}

	default:
		usage()
		os.Exit(1)
	}

	fmt.Println("Data conversion was successful.")
}
