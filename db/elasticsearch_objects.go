package db

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
	Field string            `json:"field"`
	Size  int               `json:"size"`
	Order map[string]string `json:"order"`
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
