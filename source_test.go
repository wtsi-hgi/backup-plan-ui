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
