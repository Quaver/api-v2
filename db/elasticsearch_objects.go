package db

import (
	"fmt"
	"reflect"
)

type Query struct {
	From  int                    `json:"from,omitempty"`
	Limit int                    `json:"limit,omitempty"`
	Size  int                    `json:"size,omitempty"`
	Aggs  Aggs                   `json:"aggs,omitempty"`
	Query BoolQuery              `json:"query,omitempty"`
	Sort  []map[string]SortOrder `json:"sort,omitempty"`
}

type Aggs struct {
	ByMapsetID ByMapsetID `json:"by_mapset_id"`
}

type ByMapsetID struct {
	Terms Terms       `json:"terms"`
	Aggs  GroupedHits `json:"aggs"`
}

type Terms struct {
	Field string `json:"field"`
	Size  int    `json:"size"`
}

type GroupedHits struct {
	GroupedHits TopHitsAgg `json:"grouped_hits"`
}

type TopHitsAgg struct {
	TopHits TopHits `json:"top_hits"`
}

type TopHits struct {
	Source map[string]interface{} `json:"_source"`
	Size   int                    `json:"size"`
	Sort   []map[string]SortOrder `json:"sort,omitempty"`
}

type Bool struct {
	Must               []interface{} `json:"must,omitempty"`
	Should             []interface{} `json:"should,omitempty"`
	Filter             []interface{} `json:"filter,omitempty"`
	MinimumShouldMatch int           `json:"minimum_should_match,omitempty"`
}

type BoolQuery struct {
	Bool MustBool `json:"bool"`
}

type MustBool struct {
	Must MustQuery `json:"must"`
}

type MustQuery struct {
	Bool ShouldBool `json:"bool"`
}

type ShouldBool struct {
	Should []QueryString `json:"should"`
}

type QueryString struct {
	QueryString QueryStringParams `json:"query_string"`
}

type QueryStringParams struct {
	Query           string   `json:"query"`
	Fields          []string `json:"fields"`
	DefaultOperator string   `json:"default_operator"`
	Boost           float64  `json:"boost"`
}

type TermQuery struct {
	Query Term `json:"query"`
}

type Term struct {
	Term Field `json:"term"`
}

type Field struct {
	MapsetID int `json:"mapset_id"`
}

type Sort struct {
	Field SortOrder `json:"field"`
}

type SortOrder struct {
	Order string `json:"order"`
}

func convertSortOrders(sortOrders []map[string]SortOrder) []Sort {
	var result []Sort
	for _, sortOrder := range sortOrders {
		for _, order := range sortOrder {
			result = append(result, Sort{Field: SortOrder{Order: order.Order}})
		}
	}
	return result
}

func printQueryStructure(v interface{}, indent string) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := val.Field(i)
			fmt.Printf("%s%s -> ", indent, field.Name)
			printQueryStructure(value.Interface(), indent+"  ")
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			fmt.Printf("%s%v -> ", indent, key)
			printQueryStructure(val.MapIndex(key).Interface(), indent+"  ")
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			fmt.Printf("%s[%d] -> ", indent, i)
			printQueryStructure(val.Index(i).Interface(), indent+"  ")
		}
	default:
		fmt.Printf("%v\n", v)
	}
}
