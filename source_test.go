package main

import (
	"reflect"
	"testing"
)

func TestReadAll(t *testing.T) {
	csvSource := CSVSource{path: "data/plan.csv"}

	entries, err := csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(entries) != 2 {
		t.Fatal("Number of read entries is incorrect")
	}

	expectedEntry := Entry{
		ReportingName: "test_project",
		ReportingRoot: "/path/to/project",
		Directory: "/path/to/project/input",
		Instruction: Backup,
		Requestor: "user",
		Faculty: "group",
	}

	if !reflect.DeepEqual(*entries[0], expectedEntry) {
		t.Errorf("First entry does not match expected values.\nGot %+v, expected %+v", entries[0], expectedEntry)
	}
}

func TestUpdateEntry(t *testing.T) {
	csvSource := CSVSource{path: "data/plan.csv"}

	originalEntry := Entry{
		ReportingName: "test_project",
		ReportingRoot: "/path/to/project",
		Directory: "/path/to/project/input",
		Instruction: Backup,
		Requestor: "user",
		Faculty: "group",
	}

	newEntry := originalEntry
	newEntry.ReportingName = "test_project_updated"

	err := csvSource.updateEntry(&newEntry) 
	if err != nil {
		t.Fatal(err)
	}

	entries, err := csvSource.readAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Error("CSV has more entries than expected.")
	}

	if !reflect.DeepEqual(*entries[0], newEntry) {
		t.Errorf("First entry does not match the updated entry.\nGot %+v, expected %+v", entries[0], newEntry)
	}
}
