# kendohelper v2

![kendohelper](https://img.shields.io/badge/version-2.0.0-blue.svg?style=flat)
[![kendohelper](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/muktihari/kendohelper)
![kendohelper](https://img.shields.io/badge/test%20coverage-100%25-brightgreen.svg?style=flat)

##### IMPORTANT NOTES:
1. Make sure to specify the data type in kendoGrid's schema model, especially for numbers, leave it empty may result value will be treated as a string.
2. There is no "in" operator in kendo, instead array of filters with "eq" operator will be implemented.
3. If filter has "filters" field (nested filter) that is not empty, the value inside "filters" will be used instead.
4. Working with date may need additional effort to manipulate the data, if we just pass "2019-01-01 00:00:00.000Z", we will only get exactly the date with that specific time.
5. Build for MongoDB, other DB may need some adjustments

---
## Getting Started
### The Basic Func
- Filter: 
  - ToDBOXFilter
  - ToAggregateFilter
- Sort: 
  - ToDBOXSort
  - ToAggregateSort

#### Preparing payload:
```go
payload := struct{
    ...
    Filter       kendohelper.Filter
    Sort         kendohelper.Sort
    ...
}{}

if err := k.GetPayload(&payload); err != nil {
    return err
}
```
#### Use in find query:
```go
...
query := tk.M{
    "where": payload.Filter.ToDBOXFilter(), // return *dbox.Filter
    "order": payload.Sort.ToDBOXSort(),     // return []string
}
```

#### Use in pipe command:
```go
...
pipe := []tk.M{
    tk.M{
            "$match": payload.Filter.ToAggregateFilter(), // return toolkit.M
    },
    tk.M{
            "$sort": payload.Sort.ToAggregateSort(),      // return bson.D
    }
}
```
### The Handle Func
- Filter: 
  - HandleField
  - Handle
- Sort: 
  - HandleField
  - Handle

Before calling The Basic Func, we might want to refactor the struct using these functions first.

Note: There are always two ways to reconstruct the struct, one is on JS (frontend, before the data is being sent to the server) and the second one is on Go (backend). Choose which one is making more sense to you, based on the given context.

Examples:
#### Treat the fields as lower case
```go
payload.Filter.HandleField(strings.ToLower)
payload.Sort.HandleField(strings.ToLower)
```
#### Rename specific field to something else (field aliasing)
```go
handler := func(field string) string {
    if field == "name" {
        return "fullname"
    }
    return field
}
payload.Filter.HandleField(handler)
payload.Sort.HandleField(handler)
```
#### Sometimes, we may have very dynamic columns from aggregate's result that only show when value > 0 for example.

So instead of
```go
fieldName: {$eq: 0} // will not retrieve any data cz the field itself is not exist
```
We might want to change it to
```go
fieldName: {$exists: false}
```
To do that we might want to use Handle Func
```go
payload.Filter.Handle(func(filter kendohelper.Filter) kendohelper.Filter {
    if len(filter.Filters) == 0 {
        value, ok := filter.Value.(float64)
        if ok && value == 0 && filter.Operator == "eq" {
            filter.Value = tk.M{"$exists": false}
        }
    }
    return filter
})
```

Handling this scenario on backend is making more sense because the frontend shouldn't know how the filter is actually working on backend.

Those are just sample scenarios. The Handle Func also be used as field validation such as which fields are allowed to be filtered which fields are prohibited and so on and so on.

---

## More Examples


### Filter Validation

To prevent some restricted fields from being filtered, we can just make the operator empty, unrecognized operator will make that filter unprocessed.

```go
payload.Filter.Handle(func(filter kendohelper.Filter) kendohelper.Filter {
    if filter.Field == "commission_fee" {
        filter.Operator = ""
    }
    return filter
})
```

To prevent some restricted fields from being sorted, we can make the Field and Dir to be an empty string.

```go
payload.Sort.Handle(func(sortElem kendohelper.SortElem) kendohelper.SortElem {
    if sortElem.Field == "commission_fee" {
        sortElem.Field = ""
        sortElem.Dir = ""
    }
    return sortElem
})
```

### Working with date

For example, we have field named created_at that has type of timestamp on mongo collection. To query the date that equals to 2019-01-01 between 00:00:00 to 23:59.59. We could reconstruct it like this:

#### JS side:
On parameterMap, change the data from this:
```javascript
data.filter: {
    field: "created_at",
    operator: "eq",
    value: "2019-01-01 00:00:00.000Z"
}
```
Into this:
```javascript
data.filter: {
    field: "created_at",               // will be ignored
    operator: "eq",                    // will be ignored
    value: "2019-01-01 00:00:00.000Z", // will be ignored
    filters: [{
        field: "created_at",
        operator: "gte",
        value: "2019-01-01 00:00:00.000Z"
    },
    {
        field: "created_at",
        operator: "lt",
        value: "2019-01-02 00:00:00.000Z"
    }],
    logic: "and"
}
```
#### On Go side would looks like this:
```go
Filter{
    Filters: []Filter{
        ...
        Filter{
            Field: "created_at",
            Operator: "gte",
            Value: "2019-01-01 00:00:00.000Z",
        },
        Filter{
            Field: "created_at",
            Operator: "lt",
            Value: "2019-01-02 00:00:00.000Z",
        },
    },
    Logic: "and",
}
```

How to do it on Go? We can use Handle Func:
```go
payload.Filter.Handle(func(filter kendohelper.Filter) kendohelper.Filter {
    if len(filter.Filters) == 0 {
        if filter.Field == "created_at" {
            valueStr, ok := filter.Value.(string)
            if !ok {
                return filter
            }
            t, err := time.Parse(time.RFC3339, valueStr)
            if err != nil {
                return filter
            }
            filter.Filters = []kendohelper.Filter{
                kendohelper.Filter{
                    Field:    filter.Field,
                    Operator: "gte",
                    Value:    filter.Value,
                },
                kendohelper.Filter{
                    Field:    filter.Field,
                    Operator: "lt",
                    Value:    t.AddDate(0, 0, 1).Format(time.RFC3339),
                },
            }
            filter.Logic = "and"
        }
    }
    return filter
})
```

### The Additional Func
- Filter: 
  - DeepCopyTo


Filter has field "Filters" which is slice of Filter. In go, values of slice are passed by reference, to avoid making changes to the original Filter, use DeepCopyTo instead

```go
// DON'T DO:
newFilter := payload.Filter
newFilter.HandleField(strings.ToLower) // payload.Filter will also be affected

// INSTEAD DO:
newFilter := kendofilter.Filter{}
payload.Filter.DeepCopyTo(&newFilter)
newFilter.HandleField(strings.ToLower) // completely isolated, won't affect payload.Filter

```

---

###### Thanks to [surya](https://github.com/dewanggasurya) and [radit](https://github.com/raditzlawliet) for supporting this project, any (usually unecessary) talk means a lot. :D - 2019
