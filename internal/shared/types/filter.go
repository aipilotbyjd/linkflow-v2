package types

import "time"

// FilterOperator represents a filter comparison operator
type FilterOperator string

const (
	FilterOpEqual        FilterOperator = "eq"
	FilterOpNotEqual     FilterOperator = "ne"
	FilterOpGreaterThan  FilterOperator = "gt"
	FilterOpGreaterEqual FilterOperator = "gte"
	FilterOpLessThan     FilterOperator = "lt"
	FilterOpLessEqual    FilterOperator = "lte"
	FilterOpIn           FilterOperator = "in"
	FilterOpNotIn        FilterOperator = "nin"
	FilterOpLike         FilterOperator = "like"
	FilterOpContains     FilterOperator = "contains"
	FilterOpIsNull       FilterOperator = "null"
	FilterOpIsNotNull    FilterOperator = "notnull"
)

// Filter represents a single filter condition
type Filter struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
}

// FilterSet represents a collection of filters
type FilterSet struct {
	Filters []Filter `json:"filters"`
	Logic   string   `json:"logic"` // "and" or "or"
}

func NewFilterSet() *FilterSet {
	return &FilterSet{
		Filters: make([]Filter, 0),
		Logic:   "and",
	}
}

func (f *FilterSet) Add(field string, op FilterOperator, value interface{}) *FilterSet {
	f.Filters = append(f.Filters, Filter{
		Field:    field,
		Operator: op,
		Value:    value,
	})
	return f
}

func (f *FilterSet) Equal(field string, value interface{}) *FilterSet {
	return f.Add(field, FilterOpEqual, value)
}

func (f *FilterSet) NotEqual(field string, value interface{}) *FilterSet {
	return f.Add(field, FilterOpNotEqual, value)
}

func (f *FilterSet) In(field string, values ...interface{}) *FilterSet {
	return f.Add(field, FilterOpIn, values)
}

func (f *FilterSet) Like(field string, pattern string) *FilterSet {
	return f.Add(field, FilterOpLike, pattern)
}

func (f *FilterSet) IsNull(field string) *FilterSet {
	return f.Add(field, FilterOpIsNull, nil)
}

func (f *FilterSet) IsNotNull(field string) *FilterSet {
	return f.Add(field, FilterOpIsNotNull, nil)
}

// DateRange represents a date range filter
type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

func (d DateRange) IsEmpty() bool {
	return d.From == nil && d.To == nil
}

// SearchQuery represents a full-text search query
type SearchQuery struct {
	Query   string   `json:"query"`
	Fields  []string `json:"fields"`
	Fuzzy   bool     `json:"fuzzy"`
	Prefix  bool     `json:"prefix"`
}
