package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	. "github.com/smarty/assertions"
)

func TestSplitList(t *testing.T) {
	expectedList := []string{"item1", "item2", "item3"}

	var tests = []struct {
		listToSplit, testName string
	}{
		{"item1, item2, item3", "comma separated"},
		{"item1\nitem2\nitem3", "line break seperated"},
		{"item1,\nitem2,\n,item3", "line break and comma seperated"},
		{"item1, , ,,  item2, \n,item3", "mixture of separators"},
	}

	for _, tt := range tests {
		testname := tt.testName
		t.Run(testname, func(t *testing.T) {
			splitList := splitList(tt.listToSplit)

			if !reflect.DeepEqual(splitList, expectedList) {
				t.Errorf("List was not split properly.\nGot %+v, expected %+v",
					splitList, expectedList)
			}
		})
	}
}

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
