package kendohelper

/* @Author
 * Hikmatulloh Hari Mukti <hikmatullohhari@gmail.com>
 *
 * References:
 * https://docs.telerik.com/kendo-ui/api/javascript/data/datasource/configuration/filter
 */

import (
	"time"

	"github.com/eaciit/dbox"
	"github.com/eaciit/toolkit"
)

// Filter is Kendo filter's object structure.
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
	Filters  []Filter
	Logic    string
}

var defaultDBOXFilter = &dbox.Filter{
	Field: "_id",
	Op:    dbox.FilterOpEqual,
	Value: toolkit.M{"$exists": true},
}

// DefaultDBOXFilter is the default value of ToDBOXFilter() when no filter is generated.
// It avoids (panic) nil pointer dereference or empty dbox.Filter{} as result.
// The idea is, if filter is empty, the query will continue to show data.
// You could also use this variable to check, whether to use the output as filter or not.
func DefaultDBOXFilter() *dbox.Filter {
	return defaultDBOXFilter
}

// The FilterHandleFunc type is an adapter to allow the use of the Filter's Handler
type FilterHandleFunc func(Filter) Filter

// Handle handles refactoring Filter struct
func (f *Filter) Handle(handler FilterHandleFunc) {
	if len(f.Filters) == 0 {
		*f = handler(*f)
		return
	}
	for i := range f.Filters {
		f.Filters[i].Handle(handler)
	}
}

// HandleField refactoring specifically on fields inside Filter struct
func (f *Filter) HandleField(handler func(field string) string) {
	f.Handle(func(filter Filter) Filter {
		filter.Field = handler(filter.Field)
		return filter
	})
}

// ToDBOXFilter converts Filter to *dbox.Filter{}.
// Querying a string, except for "eq" and "neq", is case-insensitive
func (f *Filter) ToDBOXFilter() *dbox.Filter {
	if len(f.Filters) == 0 {
		valueStr, ok := f.Value.(string)
		if ok {
			t, err := time.Parse(time.RFC3339, valueStr)
			if err == nil {
				f.Value = t
			}
		} else if f.Operator == "startswith" ||
			f.Operator == "doesnotstartwith" ||
			f.Operator == "contains" ||
			f.Operator == "doesnotcontain" ||
			f.Operator == "isempty" ||
			f.Operator == "isnotempty" {
			return defaultDBOXFilter
		}

		switch f.Operator {
		case "isnull":
			return dbox.Eq(f.Field, nil)
		case "isnotnull":
			return dbox.Ne(f.Field, nil)
		case "eq":
			return dbox.Eq(f.Field, f.Value)
		case "neq":
			return dbox.Ne(f.Field, f.Value)
		case "lt":
			return dbox.Lt(f.Field, f.Value)
		case "lte":
			return dbox.Lte(f.Field, f.Value)
		case "gt":
			return dbox.Gt(f.Field, f.Value)
		case "gte":
			return dbox.Gte(f.Field, f.Value)
		case "startswith":
			return dbox.Startwith(f.Field, valueStr)
		case "doesnotstartwith":
			return &dbox.Filter{
				Field: f.Field,
				Op:    dbox.FilterOpEqual,
				Value: toolkit.M{
					"$regex":   `^(?!` + valueStr + `)\w+`,
					"$options": "i",
				},
			}
		case "contains":
			return dbox.Contains(f.Field, valueStr)
		case "doesnotcontain":
			return &dbox.Filter{
				Field: f.Field,
				Op:    dbox.FilterOpEqual,
				Value: toolkit.M{
					"$regex":   `^((?!` + valueStr + `).)*$`,
					"$options": "i",
				},
			}
		case "isempty":
			return dbox.Eq(f.Field, "")
		case "isnotempty":
			return dbox.Ne(f.Field, "")
		}
		return defaultDBOXFilter
	}

	dboxFilters := []*dbox.Filter{}
	for _, filter := range f.Filters {
		dboxFilter := filter.ToDBOXFilter()
		if dboxFilter != defaultDBOXFilter {
			dboxFilters = append(dboxFilters, dboxFilter)
		}
	}
	if len(dboxFilters) == 0 {
		return defaultDBOXFilter
	}
	if f.Logic == "and" {
		return dbox.And(dboxFilters...)
	} else if f.Logic == "or" {
		return dbox.Or(dboxFilters...)
	}
	return defaultDBOXFilter
}

// ToAggregateFilter converts Filter to Mongo Pipeline $match (aggregation).
// Querying a string, except for "eq" and "neq", is case-insensitive
func (f *Filter) ToAggregateFilter() toolkit.M {
	if len(f.Filters) == 0 {
		valueStr, ok := f.Value.(string)
		if ok {
			t, err := time.Parse(time.RFC3339, valueStr)
			if err == nil {
				f.Value = t
			}
		} else if f.Operator == "startswith" ||
			f.Operator == "doesnotstartwith" ||
			f.Operator == "contains" ||
			f.Operator == "doesnotcontain" ||
			f.Operator == "isempty" ||
			f.Operator == "isnotempty" {
			return nil
		}

		switch f.Operator {
		case "isnull":
			return toolkit.M{f.Field: nil}
		case "isnotnull":
			return toolkit.M{f.Field: toolkit.M{"$ne": nil}}
		case "eq":
			return toolkit.M{f.Field: f.Value}
		case "neq":
			return toolkit.M{f.Field: toolkit.M{"$ne": f.Value}}
		case "lt":
			fallthrough
		case "lte":
			fallthrough
		case "gt":
			fallthrough
		case "gte":
			return toolkit.M{f.Field: toolkit.M{
				"$" + f.Operator: f.Value,
			}}
		case "startswith":
			return toolkit.M{f.Field: toolkit.M{
				"$regex":   `^` + valueStr,
				"$options": "i",
			}}
		case "doesnotstartwith":
			return toolkit.M{f.Field: toolkit.M{
				"$regex":   `^(?!` + valueStr + `)\w+`,
				"$options": "i",
			}}
		case "contains":
			return toolkit.M{f.Field: toolkit.M{
				"$regex":   `.*` + valueStr + `.*`,
				"$options": "i",
			}}
		case "doesnotcontain":
			return toolkit.M{f.Field: toolkit.M{
				"$regex":   `^((?!` + valueStr + `).)*$`,
				"$options": "i",
			}}
		case "isempty":
			return toolkit.M{f.Field: ""}
		case "isnotempty":
			return toolkit.M{f.Field: toolkit.M{"$ne": ""}}
		}
		return nil
	}

	matches := []toolkit.M{}
	for _, filter := range f.Filters {
		match := filter.ToAggregateFilter()
		if match != nil {
			matches = append(matches, match)
		}
	}
	if len(matches) == 0 {
		return nil
	}
	if f.Logic == "and" {
		return toolkit.M{"$and": matches}
	} else if f.Logic == "or" {
		return toolkit.M{"$or": matches}
	}
	return nil
}

// DeepCopyTo will copy filter as a branch new Filter to dest Filter
func (f *Filter) DeepCopyTo(dest *Filter) {
	*dest = *f
	if f.Filters == nil {
		return
	}
	dest.Filters = make([]Filter, len(f.Filters))
	copy(dest.Filters, f.Filters)
	for i := range f.Filters {
		f.Filters[i].DeepCopyTo(&dest.Filters[i])
	}
}
