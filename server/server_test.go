package server

import (
	"backup-plan-ui/sources"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	. "github.com/smarty/assertions"
)

func TestShowAddRowForm(t *testing.T) {
	s, _ := createServer(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actions/add", nil)

	s.ShowAddRowForm(w, r)

	body := getBodyAndCheckStatusOK(t, w)

	if ok, err := So(body, ShouldContainSubstring, "<table"); !ok {
		t.Error(err)
	}
}

func TestGetEntries(t *testing.T) {
	s, originalEntries := createServer(t)

	req := httptest.NewRequest(http.MethodGet, "/entries", nil)
	w := httptest.NewRecorder()

	s.GetEntries(w, req)

	body := getBodyAndCheckStatusOK(t, w)

	for _, entry := range originalEntries {
		if ok, err := So(string(body), ShouldContainSubstring, entry.ReportingName); !ok {
			t.Error(err)
		}
	}
}

func TestSubmitEdits(t *testing.T) {
	s, originalEntries := createServer(t)

	entryToEdit := originalEntries[0]

	tests := []struct {
		name     string
		entry    sources.Entry
		newValue string
	}{
		{
			name: "You can edit Reporting Name",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.ReportingName = "NewName"

				return entry
			}(),
			newValue: "NewName",
		},
		{
			name: "You can edit Reporting Root",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.ReportingRoot = "/new/root/to/project/dir"
				entry.Directory = "/new/root/to/project/dir/nested"

				return entry
			}(),
			newValue: "/new/root/to/project/dir",
		},
		{
			name: "You can edit Directory",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.Directory = "/some/path/to/project/dir/a/new/input"

				return entry
			}(),
			newValue: "/some/path/to/project/dir/a/new/input",
		},
		{
			name: "You can edit Instruction",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.Instruction = sources.NoBackup

				return entry
			}(),
			newValue: string(sources.NoBackup),
		},
		{
			name: "You can edit Match",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.Match = "*.csv *.txt"

				return entry
			}(),
			newValue: "*.csv *.txt",
		},
		{
			name: "You can edit Ignore",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.Instruction = sources.Backup
				entry.Ignore = "*.txt"

				return entry
			}(),
			newValue: "*.txt",
		},
		{
			name: "You can edit Requestor",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.Requestor = "NewRequestor"

				return entry
			}(),
			newValue: "NewRequestor",
		},
		{
			name: "You can edit Faculty",
			entry: func() sources.Entry {
				entry := *entryToEdit
				entry.Faculty = "NewFaculty"

				return entry
			}(),
			newValue: "NewFaculty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			form := createFormFromEntry(test.entry)
			req := makeFormRequest(form, fmt.Sprintf("/actions/submit/%d", test.entry.ID), fmt.Sprintf("%d", test.entry.ID))

			w := httptest.NewRecorder()

			s.SubmitEdits(w, req)

			body := getBodyAndCheckStatusOK(t, w)

			if ok, err := So(body, ShouldContainSubstring, test.newValue); !ok {
				t.Error(err)
			}

			changedEntry, err := s.DB.GetEntry(test.entry.ID)
			if err != nil {
				t.Error(err)
			}

			if ok, err := So(*changedEntry, ShouldResemble, test.entry); !ok {
				t.Error(err)
			}

			if ok, err := So(changedEntry, ShouldNotResemble, entryToEdit); !ok {
				t.Error(err)
			}
		})
	}
}

func createFormFromEntry(entry sources.Entry) url.Values {
	form := make(url.Values)

	form.Set(ReportingName.string(), entry.ReportingName)
	form.Set(ReportingRoot.string(), entry.ReportingRoot)
	form.Set(Directory.string(), entry.Directory)
	form.Set(Instruction.string(), string(entry.Instruction))
	form.Set(Match.string(), entry.Match)
	form.Set(Ignore.string(), entry.Ignore)
	form.Set(Requestor.string(), entry.Requestor)
	form.Set(Faculty.string(), entry.Faculty)

	return form
}

func TestValidateForm(t *testing.T) {
	exampleFormData := map[formField]string{
		ReportingName: "test_report",
		ReportingRoot: "/a/b/c/d/e",
		Directory:     "/a/b/c/d/e/f",
		Instruction:   "testInstruction",
		Match:         "",
		Ignore:        "",
		Requestor:     "test_user",
		Faculty:       "test_group",
	}

	for fieldName := range exampleFormData {
		if fieldName == Match || fieldName == Ignore {
			continue
		}

		t.Run(fmt.Sprintf("Blank %s", fieldName), func(t *testing.T) {
			data := cloneMap(exampleFormData)
			data[fieldName] = ""

			req := makeFormRequest(createFormFromMap(data), "/", "")
			errors := validateForm(req)

			if got := errors[fieldName]; got != ErrBlankInput {
				t.Errorf("Expected error for %s: %q, got: %q", fieldName, ErrBlankInput, got)
			}
		})
	}

	tests := []struct {
		name        string
		formData    map[formField]string
		KeyForErr   formField
		expectedErr string
	}{
		{
			name:        "Invalid instruction input",
			formData:    cloneAndUpdateMapValue(exampleFormData, Instruction, "invalid"),
			KeyForErr:   Instruction,
			expectedErr: ErrInvalidInstruction,
		},
		{
			name: "Ignore when instruction is not backup",
			formData: func() map[formField]string {
				data := cloneMap(exampleFormData)
				data[Instruction] = "nobackup"
				data[Ignore] = "*.txt"
				return data
			}(),
			KeyForErr:   Ignore,
			expectedErr: ErrIgnoreWithoutBackup,
		},
		{
			name:        "Reporting root doesn't start with a slash",
			formData:    cloneAndUpdateMapValue(exampleFormData, ReportingRoot, "some/dir"),
			KeyForErr:   ReportingRoot,
			expectedErr: ErrRootWithoutSlash,
		},
		{
			name:        "Reporting root not deep enough",
			formData:    cloneAndUpdateMapValue(exampleFormData, ReportingRoot, "/a/shallow/dir"),
			KeyForErr:   ReportingRoot,
			expectedErr: ErrReportingRootNotDeepEnough,
		},
		{
			name: "Directory not in Reporting root",
			formData: func() map[formField]string {
				data := cloneMap(exampleFormData)
				data[ReportingRoot] = "/some/root/parent/nested/dir"
				data[Directory] = "/some/other/root/parent/nested/dir"
				return data
			}(),
			KeyForErr:   Directory,
			expectedErr: ErrDirectoryNotInRoot,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := makeFormRequest(createFormFromMap(test.formData), "/", "")
			errors := validateForm(req)

			if got := errors[test.KeyForErr]; got != test.expectedErr {
				t.Errorf("Expected error for %s: %q, got: %q", test.KeyForErr, test.expectedErr, got)
			}
		})
	}
}

func createServer(t *testing.T) (Server, []*sources.Entry) {
	t.Helper()

	entries, dbPath := sources.CreateTestData(t)

	server := Server{
		DB:        sources.CSVSource{Path: dbPath},
		templates: template.Must(template.ParseGlob("../templates/*.html")),
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

func makeFormRequest(form url.Values, endpoint string, id string) *http.Request {
	req := httptest.NewRequest(http.MethodPut, endpoint, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_ = req.ParseForm()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req
}

func createFormFromMap(data map[formField]string) url.Values {
	form := make(url.Values)

	for key, value := range data {
		form.Set(key.string(), value)
	}

	return form
}

func cloneAndUpdateMapValue(origMap map[formField]string, key formField, value string) map[formField]string {
	newMap := cloneMap(origMap)
	newMap[key] = value

	return newMap
}

func cloneMap(original map[formField]string) map[formField]string {
	cloned := make(map[formField]string, len(original))

	for k, v := range original {
		cloned[k] = v
	}

	return cloned
}
