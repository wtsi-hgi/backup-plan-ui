package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"backup-plan-ui/internal"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <path-to-csv> <path-to-sqlite>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	csvPath := os.Args[1]
	sqlitePath := os.Args[2]

	err := internal.ConvertCsvToSqlite(csvPath, sqlitePath)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	fmt.Println("Data conversion successful.")
}
