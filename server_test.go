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

	res := w.Result()
	defer res.Body.Close()

	if ok, err := So(res.StatusCode, ShouldEqual, http.StatusOK); !ok {
		t.Error(err)
	}

	body, err := io.ReadAll(res.Body)
	if ok, err := So(err, ShouldBeNil); !ok {
		t.Fatal(err)
	}

	if ok, err := So(string(body), ShouldContainSubstring, "<table"); !ok {
		t.Error(err)
	}
}
