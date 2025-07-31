package main

import (
	"io"
	"net/http"
	"net/http/httptest"
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
