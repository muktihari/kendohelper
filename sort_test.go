package kendohelper_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/muktihari/kendohelper"
	"gopkg.in/mgo.v2/bson"
)

func TestSortHandleField(t *testing.T) {
	tt := []struct {
		name     string
		sort     kendohelper.Sort
		handler  func(string) string
		expected kendohelper.Sort
	}{
		{
			name: "lower case field",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
			handler: strings.ToLower,
			expected: kendohelper.Sort{
				kendohelper.SortElem{"name", "asc"},
				kendohelper.SortElem{"age", "desc"},
			},
		},
		{
			name: "upper case field",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
			handler: strings.ToUpper,
			expected: kendohelper.Sort{
				kendohelper.SortElem{"NAME", "asc"},
				kendohelper.SortElem{"AGE", "desc"},
			},
		},
		{
			name: "field aliasing",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"ID", "asc"},
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
			handler: func(field string) string {
				if field == "ID" {
					return "_id"
				}
				return field
			},
			expected: kendohelper.Sort{
				kendohelper.SortElem{"_id", "asc"},
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.sort.HandleField(tc.handler)
			if !reflect.DeepEqual(tc.sort, tc.expected) {
				t.Errorf("%v should be %v, got %v", tc.name, tc.expected, tc.sort)
			}
		})
	}
}

func TestSortHandle(t *testing.T) {
	tt := []struct {
		name     string
		sort     kendohelper.Sort
		handler  kendohelper.SortHandlerFunc
		expected kendohelper.Sort
	}{
		{
			name: "change Name field's dir from asc to desc",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
			handler: func(sortElem kendohelper.SortElem) kendohelper.SortElem {
				if sortElem.Field == "Name" {
					sortElem.Dir = "desc"
				}
				return sortElem
			},
			expected: kendohelper.Sort{
				kendohelper.SortElem{"Name", "desc"},
				kendohelper.SortElem{"Age", "desc"},
			},
		},
		{
			name: "field aliasing",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
			handler: func(sortElem kendohelper.SortElem) kendohelper.SortElem {
				if sortElem.Field == "Name" {
					sortElem.Field = "FullName"
				}
				return sortElem
			},
			expected: kendohelper.Sort{
				kendohelper.SortElem{"FullName", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.sort.Handle(tc.handler)
			if !reflect.DeepEqual(tc.sort, tc.expected) {
				t.Errorf("%v should %v, got %v", tc.name, tc.expected, tc.sort)
			}
		})
	}
}

func TestToDBOXSort(t *testing.T) {
	tt := []struct {
		name     string
		sort     kendohelper.Sort
		expected []string
	}{
		{
			name: "validate sort",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"", ""},
				kendohelper.SortElem{"Age", "desc"},
			},
			expected: []string{"-Age"},
		},
		{
			name: "single sort asc",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
			},
			expected: []string{"Name"},
		},
		{
			name: "single sort desc",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "desc"},
			},
			expected: []string{"-Name"},
		},
		{
			name: "multiple sort",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
			expected: []string{"Name", "-Age"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dboxSort := tc.sort.ToDBOXSort()
			if !reflect.DeepEqual(dboxSort, tc.expected) {
				t.Errorf("%v should be %v, got %v", tc.name, tc.expected, dboxSort)
			}
		})
	}
}

func TestToAggregateSort(t *testing.T) {
	tt := []struct {
		name     string
		sort     kendohelper.Sort
		expected bson.D
	}{
		{
			name: "validate sort",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"", ""},
				kendohelper.SortElem{"Age", "desc"},
			},
			expected: bson.D{{"Age", -1}},
		},
		{
			name: "single sort asc",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
			},
			expected: bson.D{{"Name", 1}},
		},
		{
			name: "single sort desc",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "desc"},
			},
			expected: bson.D{{"Name", -1}},
		},
		{
			name: "multiple sort",
			sort: kendohelper.Sort{
				kendohelper.SortElem{"Name", "asc"},
				kendohelper.SortElem{"Age", "desc"},
			},
			expected: bson.D{{"Name", 1}, {"Age", -1}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			aggrSort := tc.sort.ToAggregateSort()
			if !reflect.DeepEqual(aggrSort, tc.expected) {
				t.Errorf("%v should be %v, got %v", tc.name, tc.expected, aggrSort)
			}
		})
	}
}
