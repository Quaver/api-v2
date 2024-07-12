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
	} `json:"term"`
}

type Range struct {
	Gte   float64 `json:"gte,omitempty"`
	Lte   float64 `json:"lte,omitempty"`
	Boost float64 `json:"boost,omitempty"`
}

type RangeCustom struct {
	Range struct {
		DifficultyRating *Range `json:"difficulty_rating,omitempty"`
	} `json:"range"`
}
