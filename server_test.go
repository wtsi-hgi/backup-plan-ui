package main

import (
	"reflect"
	"testing"
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
