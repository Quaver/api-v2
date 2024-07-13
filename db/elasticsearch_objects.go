package db

type Query struct {
	From     int                    `json:"from,omitempty"`
	Size     int                    `json:"size"`
	Collapse Collapse               `json:"collapse,omitempty"`
	Query    BoolQuery              `json:"query,omitempty"`
	Sort     []map[string]SortOrder `json:"sort,omitempty"`
}

type Collapse struct {
	Field     string    `json:"field,omitempty"`
	InnerHits InnerHits `json:"inner_hits,omitempty"`
}

type InnerHits struct {
	Name string      `json:"name,omitempty"`
	Size int         `json:"size,omitempty"`
	Sort interface{} `json:"sort,omitempty"`
}

type BoolQuery struct {
	BoolQuery struct {
		Must               []interface{} `json:"must,omitempty"`
		Should             []interface{} `json:"should,omitempty"`
		Filter             []interface{} `json:"filter,omitempty"`
		MinimumShouldMatch int           `json:"minimum_should_match,omitempty"`
	} `json:"bool,omitempty"`
}

type QueryString struct {
	QueryString struct {
		Query           string   `json:"query"`
		Fields          []string `json:"fields"`
		DefaultOperator string   `json:"default_operator"`
		Boost           float64  `json:"boost"`
	} `json:"query_string"`
}

func NewQueryString(query string, fields []string, defaultOperator string, boost float64) QueryString {
	qs := QueryString{}

	qs.QueryString.Query = query
	qs.QueryString.Fields = fields
	qs.QueryString.DefaultOperator = defaultOperator
	qs.QueryString.Boost = boost

	return qs
}

type SortOrder struct {
	Order string `json:"order"`
}

type Term struct {
	Value interface{} `json:"value"`
	Boost float64     `json:"boost"`
}

type TermCustom struct {
	Term struct {
		GameMode     *Term `json:"game_mode,omitempty"`
		RankedStatus *Term `json:"ranked_status,omitempty"`
		Explicit     *Term `json:"explicit,omitempty"`
	} `json:"term"`
}

type Range struct {
	Gte   interface{} `json:"gte,omitempty"`
	Lte   interface{} `json:"lte,omitempty"`
	Boost interface{} `json:"boost,omitempty"`
}

type RangeCustom struct {
	Range map[string]Range `json:"range,omitempty"`
}

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func addRangeQuery[T Number](boolQuery *BoolQuery, field string, min T, max T) {
	if min != 0 || max != 0 {
		rangeCustom := RangeCustom{
			Range: map[string]Range{
				field: {
					Gte: min,
					Lte: max,
				},
			},
		}
		boolQuery.BoolQuery.Must = append(boolQuery.BoolQuery.Must, rangeCustom)
	}
}
