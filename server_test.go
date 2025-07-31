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

func TestValidateForm(t *testing.T) {
	exampleFormData := map[string]string{
		"ReportingName": "test_report",
		"ReportingRoot": "a/",
		"Directory":     "a/b/c/d/e`",
		"Instruction":   "testInstruction",
		"Match":         "",
		"Ignore":        "",
		"Requestor":     "test_user",
		"Faculty":       "test_group",
	}

	makeRequest := func(data map[string]string) *http.Request {
		form := createForm(data)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_ = req.ParseForm()

		return req
	}

	for fieldName := range exampleFormData {
		if fieldName == "Match" || fieldName == "Ignore" {
			continue
		}

		t.Run(fmt.Sprintf("Blank %s", fieldName), func(t *testing.T) {
			data := cloneMap(exampleFormData)
			data[fieldName] = ""

			req := makeRequest(data)
			errors, err := validateForm(req)
			if err != nil {
				t.Fatal(err)
			}

			if got := errors[fieldName]; got != ErrBlankInput {
				t.Errorf("Expected error for %s: %q, got: %q", fieldName, ErrBlankInput, got)
			}
		})
	}

	tests := []struct {
		name        string
		formData    map[string]string
		KeyForErr   string
		expectedErr string
	}{
		{
			name:        "invalid instruction input",
			formData:    cloneAndUpdateMapValue(exampleFormData, "Instruction", "invalid"),
			KeyForErr:   "Instruction",
			expectedErr: ErrInvalidInstruction,
		},
		{
			name: "Ignore when instruction is not backup",
			formData: func() map[string]string {
				data := cloneMap(exampleFormData)
				data["Instruction"] = "nobackup"
				data["Ignore"] = "*.txt"
				return data
			}(),
			KeyForErr:   "Ignore",
			expectedErr: ErrIgnoreWithoutBackup,
		},
		{
			name:        "Directory not deep enough",
			formData:    cloneAndUpdateMapValue(exampleFormData, "Directory", "a/shallow/dir"),
			KeyForErr:   "Directory",
			expectedErr: ErrDirectoryNotDeepEnough,
		},
		{
			name: "Directory not in Reporting root",
			formData: func() map[string]string {
				data := cloneMap(exampleFormData)
				data["ReportingRoot"] = "some/parent/"
				data["Directory"] = "some/other/parent/nested/dir"
				return data
			}(),
			KeyForErr:   "Directory",
			expectedErr: ErrDirectoryNotInRoot,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := makeRequest(test.formData)
			errors, err := validateForm(req)
			if err != nil {
				t.Fatal(err)
			}

			if got := errors[test.KeyForErr]; got != test.expectedErr {
				t.Errorf("Expected error for %s: %q, got: %q", test.KeyForErr, test.expectedErr, got)
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

func cloneAndUpdateMapValue(origMap map[string]string, key, value string) map[string]string {
	newMap := cloneMap(origMap)
	newMap[key] = value

	return newMap
}

func cloneMap(original map[string]string) map[string]string {
	cloned := make(map[string]string)
	for k, v := range original {
		cloned[k] = v
	}

	return cloned
}
