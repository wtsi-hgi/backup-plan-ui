package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wtsi-hgi/backup-plan-ui/converter"
	"github.com/wtsi-hgi/backup-plan-ui/sources"
	"github.com/wtsi-hgi/backup-plan-ui/sourcesNewSchema" //nolint:goimports
)

var (
	dropTable       bool
	sourceTableName string
	dirsTableName   string
	rulesTableName  string
)

func usage() {
	prog := filepath.Base(os.Args[0])

	fmt.Println("Convert data from old schema to new schema within MySQL database.")
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [--source-table entries] [--dirs-table dirs] [--roles-table roles] [--replace]\n", prog)

	vars := "SOURCE_MYSQL_HOST, SOURCE_MYSQL_PORT, SOURCE_MYSQL_USER, SOURCE_MYSQL_PASS, SOURCE_MYSQL_DATABASE"

	fmt.Println("\nEnvironment (source mysql): " + vars)

	vars = "TARGET_MYSQL_HOST, TARGET_MYSQL_PORT, TARGET_MYSQL_USER, TARGET_MYSQL_PASS, TARGET_MYSQL_DATABASE"

	fmt.Println("Environment (target mysql): " + vars)
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
}

func init() {
	flag.BoolVar(&dropTable, "replace", false, "Remove existing data before inserting new data")
	flag.StringVar(&sourceTableName, "source-table", sources.DefaultTableName, "Name of table to copy data from")
	flag.StringVar(&dirsTableName, "dirs-table", sourcesnewschema.DefaultDirectoriesTableName,
		"Name of table to insert directories to")
	flag.StringVar(&rulesTableName, "roles-table", sourcesnewschema.DefaultRulesTableName,
		"Name of table to insert rules to")

	flag.Usage = usage
}

func main() {
	flag.Parse()

	cfgSource := sources.MySQLConfig{
		Host:     os.Getenv("SOURCE_MYSQL_HOST"),
		Port:     os.Getenv("SOURCE_MYSQL_PORT"),
		User:     os.Getenv("SOURCE_MYSQL_USER"),
		Password: os.Getenv("SOURCE_MYSQL_PASS"),
		Database: os.Getenv("SOURCE_MYSQL_DATABASE"),
	}

	cfgTarget := sourcesnewschema.MySQLConfig{
		Host:     os.Getenv("TARGET_MYSQL_HOST"),
		Port:     os.Getenv("TARGET_MYSQL_PORT"),
		User:     os.Getenv("TARGET_MYSQL_USER"),
		Password: os.Getenv("TARGET_MYSQL_PASS"),
		Database: os.Getenv("TARGET_MYSQL_DATABASE"),
	}

	err := converter.ConvertSchema(cfgSource, sourceTableName, cfgTarget, dirsTableName, rulesTableName, dropTable)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	log.Println("Conversion successful")
}
