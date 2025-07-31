package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/smarty/assertions"
)

func TestShowAddRowForm(t *testing.T) {
	s := server{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actions/add", nil)

	s.showAddRowForm(w, r)

	body := getBodyAndCheckStatusOK(t, w)

	if ok, err := So(body, ShouldContainSubstring, "<table"); !ok {
		t.Error(err)
	}
}

func TestGetEntries(t *testing.T) {
	s, originalEntries := createServer(t)

	req := httptest.NewRequest(http.MethodGet, "/entries", nil)
	w := httptest.NewRecorder()

	s.getEntries(w, req)

	body := getBodyAndCheckStatusOK(t, w)

	for _, entry := range originalEntries {
		if ok, err := So(string(body), ShouldContainSubstring, entry.ReportingName); !ok {
			t.Error(err)
		}
	}
}

func TestAddNewEntry(t *testing.T) {
	s, _ := createServer(t)

	exampleFormData := map[string]string{
		"ReportingName": "test_report",
		"ReportingRoot": "test_root",
		"Directory":     "some/nested/dir",
		"Instruction":   "testInstruction",
		"Match":         "",
		"Ignore":        "",
		"Requestor":     "test_user",
		"Faculty":       "test_group",
	}

	createAndAddForm := func(map[string]string) string {
		form := createForm(exampleFormData)

		req := httptest.NewRequest(http.MethodPut, "/actions/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		s.addNewEntry(w, req)

		return getBodyAndCheckStatusOK(t, w)
	}

	for fieldName := range exampleFormData {
		if fieldName == "Match" || fieldName == "Ignore" {
			continue
		}

		t.Run(fmt.Sprintf("Blank %s", fieldName), func(t *testing.T) {
			exampleFormData[fieldName] = ""

			body := createAndAddForm(exampleFormData)

			if ok, err := So(body, ShouldContainSubstring, "You cannot leave this field blank"); !ok {
				t.Error(err)
			}
		})
	}
}

func createServer(t *testing.T) (server, []*Entry) {
	t.Helper()

	entries, dbPath := createTestData(t)

	server := server{
		db: CSVSource{dbPath},
	}

	return server, entries
}

func getBodyAndCheckStatusOK(t *testing.T, w *httptest.ResponseRecorder) string {
	res := w.Result()
	defer res.Body.Close()

	if ok, err := So(res.StatusCode, ShouldEqual, http.StatusOK); !ok {
		t.Error(err)
	}

	body, err := io.ReadAll(res.Body)
	if ok, err := So(err, ShouldBeNil); !ok {
		t.Fatal(err)
	}

	return string(body)
}

func createForm(data map[string]string) url.Values {
	form := make(url.Values)

	for key, value := range data {
		form.Set(key, value)
	}

	return form
}

func updateDataValue(origMap map[string]string, key, value string) map[string]string {
	origMap[key] = value

	return origMap
}
