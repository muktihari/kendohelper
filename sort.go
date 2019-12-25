package kendohelper

/* @Author
 * Hikmatulloh Hari Mukti <hikmatullohhari@gmail.com>
 *
 * References:
 * https://docs.telerik.com/kendo-ui/api/javascript/data/datasource/configuration/sort
 */

import (
	"gopkg.in/mgo.v2/bson"
)

// SortElem is element of Kendo's sort array
type SortElem struct {
	Field string
	Dir   string
}

// Sort is Kendo sort's array structure.
type Sort []SortElem

// The SortHandlerFunc type is an adapter to allow the use of the Sort's Handler
type SortHandlerFunc func(SortElem) SortElem

// Handle handles refactoring the struct before execute the ToDBOXSort or ToAggregateSort func
func (s *Sort) Handle(handler SortHandlerFunc) {
	for i, sortElem := range *s {
		(*s)[i] = handler(sortElem)
	}
}

// HandleField handles refactoring specifically on fields inside Sort struct
func (s *Sort) HandleField(handler func(field string) string) {
	s.Handle(func(sortElem SortElem) SortElem {
		sortElem.Field = handler(sortElem.Field)
		return sortElem
	})
}

// ToDBOXSort converts Sort to []string of sort by ordered field
func (s *Sort) ToDBOXSort() []string {
	sort := []string{}
	for _, v := range *s {
		if v.Dir != "asc" && v.Dir != "desc" {
			continue
		}
		if v.Dir == "desc" {
			v.Field = "-" + v.Field
		}
		sort = append(sort, v.Field)
	}
	return sort
}

// ToAggregateSort converts Sort to bson.D (Ordered Map) used in Mongo Pipeline $sort (aggregation)
func (s *Sort) ToAggregateSort() bson.D {
	sort := bson.D{}
	for _, v := range *s {
		if v.Dir != "asc" && v.Dir != "desc" {
			continue
		}
		dir := 1
		if v.Dir == "desc" {
			dir = -1
		}
		sort = append(sort, bson.DocElem{
			Name:  v.Field,
			Value: dir,
		})
	}
	return sort
}

func (s *Sort) DeepCopy() Sort {
	sort := make(Sort, len(*s))
	for i, v := range *s {
		sort[i] = v
	}
	return sort
}

func (s *Sort) HasField(fields ...string) bool {
	for _, v := range *s {
		for _, field := range fields {
			if v.Field == field {
				return true
			}
		}
	}
	return false
}
