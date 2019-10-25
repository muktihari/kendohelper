package kendohelper_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/eaciit/dbox"
	"github.com/eaciit/toolkit"
	"github.com/muktihari/kendohelper"
)

func TestFilterHandleField(t *testing.T) {
	tt := []struct {
		name     string
		filter   kendohelper.Filter
		handler  func(string) string
		expected kendohelper.Filter
	}{
		{
			name: "lower case field",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
				kendohelper.Filter{"Age", "eq", 25, nil, ""},
			}, "and"},
			handler: strings.ToLower,
			expected: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"name", "eq", "Hari", nil, ""},
				kendohelper.Filter{"age", "eq", 25, nil, ""},
			}, "and"},
		},
		{
			name: "upper case field",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
				kendohelper.Filter{"Age", "eq", 25, nil, ""},
			}, "and"},
			handler: strings.ToUpper,
			expected: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"NAME", "eq", "Hari", nil, ""},
				kendohelper.Filter{"AGE", "eq", 25, nil, ""},
			}, "and"},
		},
		{
			name: "field aliasing",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"ID", "eq", "Hari", nil, ""},
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
				kendohelper.Filter{"Age", "eq", 25, nil, ""},
			}, "and"},
			handler: func(field string) string {
				if field == "ID" {
					return "_id"
				}
				return field
			},
			expected: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"_id", "eq", "Hari", nil, ""},
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
				kendohelper.Filter{"Age", "eq", 25, nil, ""},
			}, "and"},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.filter.HandleField(tc.handler)
			if !reflect.DeepEqual(tc.filter, tc.expected) {
				t.Errorf("%v should %v, got %v", tc.name, tc.expected, tc.filter)
			}
		})
	}
}

func TestFilterHandle(t *testing.T) {
	tt := []struct {
		name     string
		filter   kendohelper.Filter
		handler  kendohelper.FilterHandleFunc
		expected kendohelper.Filter
	}{
		{
			name: "change Name field's operator from eq to neq",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
				kendohelper.Filter{"Age", "eq", 25, nil, ""},
			}, "and"},
			handler: func(filter kendohelper.Filter) kendohelper.Filter {
				if filter.Field == "Name" {
					filter.Operator = "neq"
				}
				return filter
			},
			expected: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "neq", "Hari", nil, ""},
				kendohelper.Filter{"Age", "eq", 25, nil, ""},
			}, "and"},
		},
		{
			name: "change created_at field operator from 'exactly eq to specific time' into 'between range of time' (gte and lt)",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"created_at", "eq", "2019-01-01T00:00:00Z", nil, ""},
			}, "and"},
			handler: func(filter kendohelper.Filter) kendohelper.Filter {
				if filter.Field == "created_at" {
					valueStr, ok := filter.Value.(string)
					if !ok {
						t.Fatalf("value is not a string")
					}
					createdAt, err := time.ParseInLocation(time.RFC3339, valueStr, time.UTC)
					if err != nil {
						t.Fatalf("cant parse to time, err: %v", err)
					}
					filter.Filters = []kendohelper.Filter{
						kendohelper.Filter{filter.Field, "gte", filter.Value, nil, ""},
						kendohelper.Filter{filter.Field, "lt", createdAt.AddDate(0, 0, 1).Format(time.RFC3339), nil, ""},
					}
					filter.Logic = "and"
				}
				return filter
			},
			expected: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"created_at", "eq", "2019-01-01T00:00:00Z", []kendohelper.Filter{
					kendohelper.Filter{"created_at", "gte", "2019-01-01T00:00:00Z", nil, ""},
					kendohelper.Filter{"created_at", "lt", "2019-01-02T00:00:00Z", nil, ""},
				}, "and"},
			}, "and"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.filter.Handle(tc.handler)
			if !reflect.DeepEqual(tc.filter, tc.expected) {
				t.Errorf("%v should %v, got %v", tc.name, tc.expected, tc.filter)
			}
		})
	}
}

func TestToDBOXFilter(t *testing.T) {
	tt := []struct {
		name     string
		filter   kendohelper.Filter
		expected *dbox.Filter
	}{
		{
			name: "operator unrecognized",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "ne", "Hari", nil, ""},
			}, "and"},
			expected: kendohelper.DefaultDBOXFilter(),
		},
		{
			name: "no logic declared",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
			}, ""},
			expected: kendohelper.DefaultDBOXFilter(),
		},
		{
			name: "value is not a string but using string's operator",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "startswith", 25, nil, ""},
			}, "and"},
			expected: kendohelper.DefaultDBOXFilter(),
		},
		{
			name: "lt or gt",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "lt", 25, nil, ""},
				kendohelper.Filter{"Age", "gt", 25, nil, ""},
			}, "or"},
			expected: dbox.Or(dbox.Lt("Age", 25), dbox.Gt("Age", 25)),
		},
		{
			name: "isnull",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isnull", nil, nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Eq("Name", nil)),
		},
		{
			name: "isnotnull",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isnotnull", nil, nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Ne("Name", nil)),
		},
		{
			name: "eq",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Eq("Name", "Hari")),
		},
		{
			name: "neq",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "neq", "Hari", nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Ne("Name", "Hari")),
		},
		{
			name: "lt",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "lt", 25, nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Lt("Age", 25)),
		},
		{
			name: "lte",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "lte", 25, nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Lte("Age", 25)),
		},
		{
			name: "gt",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "gt", 25, nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Gt("Age", 25)),
		},
		{
			name: "gte",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "gte", 25, nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Gte("Age", 25)),
		},
		{
			name: "startswith",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "startswith", "H", nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Startwith("Name", "H")),
		},
		{
			name: "doesnotstartwith",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "doesnotstartwith", "H", nil, ""},
			}, "and"},
			expected: dbox.And(&dbox.Filter{
				Field: "Name",
				Op:    dbox.FilterOpEqual,
				Value: toolkit.M{
					"$regex":   `^(?!H)\w+`,
					"$options": "i",
				},
			}),
		},
		{
			name: "contains",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "contains", "H", nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Contains("Name", "H")),
		},
		{
			name: "doesnotcontain",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "doesnotcontain", "H", nil, ""},
			}, "and"},
			expected: dbox.And(&dbox.Filter{
				Field: "Name",
				Op:    dbox.FilterOpEqual,
				Value: toolkit.M{
					"$regex":   `^((?!H).)*$`,
					"$options": "i",
				},
			}),
		},
		{
			name: "isempty",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isempty", "", nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Eq("Name", "")),
		},
		{
			name: "isnotempty",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isnotempty", "", nil, ""},
			}, "and"},
			expected: dbox.And(dbox.Ne("Name", "")),
		},
		{
			name: "working with date",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"created_at", "eq", "2019-01-01T00:00:00Z", []kendohelper.Filter{
					kendohelper.Filter{"created_at", "gte", "2019-01-01T00:00:00Z", nil, ""},
					kendohelper.Filter{"created_at", "lt", "2019-01-02T00:00:00Z", nil, ""},
				}, "and"},
			}, "and"},
			expected: dbox.And(
				dbox.And(
					dbox.Gte("created_at", time.Date(2019, 01, 01, 00, 00, 00, 00, time.UTC)),
					dbox.Lt("created_at", time.Date(2019, 01, 02, 00, 00, 00, 00, time.UTC)),
				),
			),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dboxFilter := tc.filter.ToDBOXFilter()
			if !reflect.DeepEqual(dboxFilter, tc.expected) {
				t.Errorf("%v should be %v, got %v", tc.name, tc.expected, dboxFilter)
			}
		})
	}
}

func TestToAggregateFilter(t *testing.T) {
	tt := []struct {
		name     string
		filter   kendohelper.Filter
		expected toolkit.M
	}{
		{
			name: "operator unrecognized",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "ne", "Hari", nil, ""},
			}, "and"},
			expected: nil,
		},
		{
			name: "no logic declared",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
			}, ""},
			expected: nil,
		},
		{
			name: "value is not a string but using string's operator",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "startswith", 25, nil, ""},
			}, "and"},
			expected: nil,
		},
		{
			name: "lt or gt",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "lt", 25, nil, ""},
				kendohelper.Filter{"Age", "gt", 25, nil, ""},
			}, "or"},
			expected: toolkit.M{"$or": []toolkit.M{
				toolkit.M{"Age": toolkit.M{"$lt": 25}},
				toolkit.M{"Age": toolkit.M{"$gt": 25}},
			}},
		},
		{
			name: "isnull",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isnull", nil, nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": nil},
			}},
		},
		{
			name: "isnotnull",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isnotnull", nil, nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": toolkit.M{"$ne": nil}},
			}},
		},
		{
			name: "eq",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "eq", "Hari", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": "Hari"},
			}},
		},
		{
			name: "neq",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "neq", "Hari", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": toolkit.M{"$ne": "Hari"}},
			}},
		},
		{
			name: "lt",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "lt", 25, nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Age": toolkit.M{"$lt": 25}},
			}},
		},
		{
			name: "lte",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "lte", 25, nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Age": toolkit.M{"$lte": 25}},
			}},
		},
		{
			name: "gt",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "gt", 25, nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Age": toolkit.M{"$gt": 25}},
			}},
		},
		{
			name: "gte",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Age", "gte", 25, nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Age": toolkit.M{"$gte": 25}},
			}},
		},
		{
			name: "startswith",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "startswith", "H", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": toolkit.M{
					"$regex":   `^` + "H",
					"$options": "i",
				}},
			}},
		},
		{
			name: "doesnotstartwith",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "doesnotstartwith", "H", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": toolkit.M{
					"$regex":   `^(?!H)\w+`,
					"$options": "i",
				}},
			}},
		},
		{
			name: "contains",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "contains", "H", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": toolkit.M{
					"$regex":   `.*H.*`,
					"$options": "i",
				}},
			}},
		},
		{
			name: "doesnotcontain",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "doesnotcontain", "H", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": toolkit.M{
					"$regex":   `^((?!H).)*$`,
					"$options": "i",
				}},
			}},
		},
		{
			name: "isempty",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isempty", "", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": ""},
			}},
		},
		{
			name: "isnotempty",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"Name", "isnotempty", "", nil, ""},
			}, "and"},
			expected: toolkit.M{"$and": []toolkit.M{
				toolkit.M{"Name": toolkit.M{"$ne": ""}},
			}},
		},
		{
			name: "working with date",
			filter: kendohelper.Filter{"", "", "", []kendohelper.Filter{
				kendohelper.Filter{"created_at", "eq", "2019-01-01T00:00:00Z", []kendohelper.Filter{
					kendohelper.Filter{"created_at", "gte", "2019-01-01T00:00:00Z", nil, ""},
					kendohelper.Filter{"created_at", "lt", "2019-01-02T00:00:00Z", nil, ""},
				}, "and"},
			}, "and"},
			expected: toolkit.M{
				"$and": []toolkit.M{
					toolkit.M{
						"$and": []toolkit.M{
							toolkit.M{"created_at": toolkit.M{"$gte": time.Date(2019, 01, 01, 00, 00, 00, 00, time.UTC)}},
							toolkit.M{"created_at": toolkit.M{"$lt": time.Date(2019, 01, 02, 00, 00, 00, 00, time.UTC)}},
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			aggrFilter := tc.filter.ToAggregateFilter()
			if !reflect.DeepEqual(aggrFilter, tc.expected) {
				t.Errorf("%v should be %v, got %v", tc.name, tc.expected, aggrFilter)
			}
		})
	}
}
