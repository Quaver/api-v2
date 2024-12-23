package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/sliceutil"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/sirupsen/logrus"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type ElasticMapsetSearchOptions struct {
	Search            string               `form:"search" json:"search"`
	RankedStatus      []enums.RankedStatus `form:"ranked_status" json:"ranked_status"`
	RankedStatusArray []enums.RankedStatus `form:"ranked_status[]" json:"ranked_status[]"`
	Mode              []enums.GameMode     `form:"mode" json:"mode"`
	ModeArray         []enums.GameMode     `form:"mode[]" json:"mode[]"`
	Page              int                  `form:"page" json:"page"`
	Limit             int                  `form:"limit" json:"limit"`
	SortBy            string               `form:"sort_by" json:"sort_by"`
	SortOrder         string               `form:"sort_order" json:"sort_order"`

	MinDifficultyRating float64 `form:"min_difficulty_rating" json:"min_difficulty_rating"`
	MaxDifficultyRating float64 `form:"max_difficulty_rating" json:"max_difficulty_rating"`
	MinBPM              float32 `form:"min_bpm" json:"min_bpm"`
	MaxBPM              float32 `form:"max_bpm" json:"max_bpm"`
	MinLength           float32 `form:"min_length" json:"min_length"`
	MaxLength           float32 `form:"max_length" json:"max_length"`
	MinLongNotePercent  float32 `form:"min_long_note_percent" json:"min_long_note_percent"`
	MaxLongNotePercent  float32 `form:"max_long_note_percent" json:"max_long_note_percent"`
	MinPlayCount        int64   `form:"min_play_count" json:"min_play_count"`
	MaxPlayCount        int64   `form:"max_play_count" json:"max_play_count"`
	MinCombo            int64   `form:"min_combo" json:"min_combo"`
	MaxCombo            int64   `form:"max_combo" json:"max_combo"`
	MinDateSubmitted    int64   `form:"min_date_submitted" json:"min_date_submitted"`
	MaxDateSubmitted    int64   `form:"max_date_submitted" json:"max_date_submitted"`
	MinLastUpdated      int64   `form:"min_last_updated" json:"min_last_updated"`
	MaxLastUpdated      int64   `form:"max_last_updated" json:"max_last_updated"`
	Explicit            bool    `form:"show_explicit" json:"show_explicit"`
	IsClanRanked        bool    `form:"is_clan_ranked" json:"is_clan_ranked"`
}

func NewElasticMapsetSearchOptions() *ElasticMapsetSearchOptions {
	return &ElasticMapsetSearchOptions{
		Search:              "",
		Limit:               50,
		MinDifficultyRating: 0,
		MaxDifficultyRating: math.MaxInt32,
		MinBPM:              0,
		MaxBPM:              math.MaxInt32,
		MinLength:           0,
		MaxLength:           math.MaxInt32,
		MinLongNotePercent:  0,
		MaxLongNotePercent:  math.MaxInt32,
		MinPlayCount:        0,
		MaxPlayCount:        math.MaxInt32,
		MinCombo:            0,
		MaxCombo:            math.MaxInt32,
		MinDateSubmitted:    0,
		MaxDateSubmitted:    math.MaxInt64,
		MinLastUpdated:      0,
		MaxLastUpdated:      math.MaxInt64,
		Explicit:            false,
		SortBy:              "date_last_updated",
		SortOrder:           "desc",
	}
}

func (options *ElasticMapsetSearchOptions) BindAndValidate() {
	if options.Limit > 50 {
		options.Limit = 50
	}

	options.Mode = sliceutil.BindMultiple(options.Mode, options.ModeArray, []enums.GameMode{
		enums.GameModeKeys4,
		enums.GameModeKeys7,
	})

	options.RankedStatus = sliceutil.BindMultiple(options.RankedStatus, options.RankedStatusArray, []enums.RankedStatus{
		enums.RankedStatusRanked,
	})
}

type ElasticMap struct {
	*MapQua
	Explicit        bool   `json:"explicit"`
	PackageMD5      string `json:"package_md5"`
	DateSubmitted   int64  `json:"date_submitted"`
	DateLastUpdated int64  `json:"date_last_updated"`
}

type ElasticHits struct {
	Hits struct {
		Hits []struct {
			InnerHits struct {
				MostRelevant struct {
					Hits struct {
						Hits []struct {
							Source ElasticMap `json:"_source"`
						} `json:"hits"`
					} `json:"hits"`
				} `json:"most_relevant"`
			} `json:"inner_hits"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		Distinct struct {
			Total int `json:"value"`
		} `json:"distinct_mapset_ids"`
	} `json:"aggregations"`
}

// IndexElasticSearchMapset Indexes an individual mapset in elastic
func IndexElasticSearchMapset(mapset Mapset) error {
	if err := DeleteElasticSearchMapset(mapset.Id); err != nil {
		return err
	}

	for _, mapQua := range mapset.Maps {
		elasticMap := ElasticMap{
			MapQua:          mapQua,
			DateSubmitted:   mapset.DateSubmitted,
			DateLastUpdated: mapset.DateLastUpdated,
			Explicit:        mapset.IsExplicit,
		}

		data, err := json.Marshal(&elasticMap)

		if err != nil {
			return err
		}

		resp, err := ElasticSearch.Create(elasticMapSearchIndex,
			fmt.Sprintf("%v", mapQua.Id), bytes.NewReader(data))

		if err != nil {
			return err
		}

		defer resp.Body.Close()
	}

	return nil
}

// UpdateElasticSearchMapset Updates an individual mapset in elastic
func UpdateElasticSearchMapset(mapset Mapset) error {
	for _, mapQua := range mapset.Maps {
		elasticMap := ElasticMap{
			MapQua:          mapQua,
			DateSubmitted:   mapset.DateSubmitted,
			DateLastUpdated: mapset.DateLastUpdated,
			Explicit:        mapset.IsExplicit,
		}

		data, err := json.Marshal(&elasticMap)

		if err != nil {
			return err
		}

		resp, err := ElasticSearch.Update(elasticMapsetIndex, fmt.Sprintf("%v", mapQua.Id), bytes.NewReader(data))

		if err != nil {
			return err
		}

		resp.Body.Close()
	}

	return nil
}

// DeleteElasticSearchMapset Deletes an individual mapset in elastic
func DeleteElasticSearchMapset(id int) error {
	queryMap := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"mapset_id": id,
			},
		},
	}

	queryJSON, err := json.Marshal(queryMap)

	if err != nil {
		return errors.New(fmt.Sprintf("Error marshalling query to JSON: %s", err))
	}

	resp, err := ElasticSearch.DeleteByQuery([]string{elasticMapSearchIndex}, strings.NewReader(string(queryJSON)))

	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

// IndexAllElasticSearchMapsets Indexes all mapsets in the DB in ElasticSearch
func IndexAllElasticSearchMapsets(deletePrevious bool) error {
	if deletePrevious {
		if err := DeleteElasticIndices(elasticMapSearchIndex); err != nil {
			return err
		}
	}

	mapsets, err := GetAllMapsets()

	if err != nil {
		return err
	}

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         elasticMapSearchIndex,
		Client:        ElasticSearch,
		NumWorkers:    5,
		FlushBytes:    int(5e+6),
		FlushInterval: 30 * time.Second,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("Error creating the indexer: %s", err))
	}

	// Put all mapsets into the task queue
	for _, mapset := range mapsets {
		for _, mapQua := range mapset.Maps {
			elasticMap := ElasticMap{
				MapQua:          mapQua,
				PackageMD5:      mapset.PackageMD5,
				DateSubmitted:   mapset.DateSubmitted,
				DateLastUpdated: mapset.DateLastUpdated,
				Explicit:        mapset.IsExplicit,
			}

			data, err := json.Marshal(&elasticMap)

			if err != nil {
				return err
			}

			err = bi.Add(
				context.Background(),
				esutil.BulkIndexerItem{
					Action:     "index",
					DocumentID: strconv.Itoa(mapQua.Id),
					Body:       bytes.NewReader(data),
					OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
						if err != nil {
							logrus.Errorf("ERROR: %s", err)
						} else {
							logrus.Errorf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
						}
					},
				},
			)

			if err != nil {
				return errors.New(fmt.Sprintf("Unexpected error:: %s", err))
			}
		}
	}

	if err := bi.Close(context.Background()); err != nil {
		logrus.Fatalf("Unexpected error: %s", err)
	}

	biStats := bi.Stats()

	logrus.Info("Successfully Indexed: ", biStats.NumFlushed)
	logrus.Info("Failed: ", biStats.NumFailed)
	return nil
}

// SearchElasticMapsets Searches ElasticSearch for mapsets
func SearchElasticMapsets(options *ElasticMapsetSearchOptions) ([]*Mapset, int, error) {
	boolQuery := BoolQuery{}

	if options.Search != "" {
		boolQuerySearch := BoolQuery{}

		qs := NewQueryString(options.Search, []string{"title", "artist"}, "OR", 1.0)
		qs2 := NewQueryString(options.Search, []string{"source", "creator_username", "difficulty_name"}, "OR", 0.8)

		m := map[string]interface{}{
			"match": map[string]interface{}{
				"tags": map[string]interface{}{
					"query": options.Search,
					"boost": 0.2,
				},
			},
		}

		boolQuerySearch.BoolQuery.Should = append(boolQuerySearch.BoolQuery.Should, qs, qs2, m)
		boolQuery.BoolQuery.Must = append(boolQuery.BoolQuery.Must, boolQuerySearch)
	}

	if options.Mode != nil {
		boolQueryMode := BoolQuery{}

		for _, mode := range options.Mode {
			termCustom := TermCustom{}
			termCustom.Term.GameMode = &Term{
				Value: mode,
				Boost: 1.0,
			}

			boolQueryMode.BoolQuery.Should = append(boolQueryMode.BoolQuery.Should, termCustom)
		}

		boolQuery.BoolQuery.Must = append(boolQuery.BoolQuery.Must, boolQueryMode)
	}

	if options.RankedStatus != nil {
		boolQueryStatus := BoolQuery{}

		for _, status := range options.RankedStatus {
			termCustom := TermCustom{}
			termCustom.Term.RankedStatus = &Term{
				Value: status,
				Boost: 1.0,
			}

			boolQueryStatus.BoolQuery.Should = append(boolQueryStatus.BoolQuery.Should, termCustom)
		}

		boolQuery.BoolQuery.Must = append(boolQuery.BoolQuery.Must, boolQueryStatus)
	}

	addRangeQuery(&boolQuery, "difficulty_rating", options.MinDifficultyRating, options.MaxDifficultyRating)
	addRangeQuery(&boolQuery, "bpm", options.MinBPM, options.MaxBPM)
	addRangeQuery(&boolQuery, "length", options.MinLength, options.MaxLength)
	addRangeQuery(&boolQuery, "long_note_percentage", options.MinLongNotePercent, options.MaxLongNotePercent)
	addRangeQuery(&boolQuery, "play_count", options.MinPlayCount, options.MaxPlayCount)
	addRangeQuery(&boolQuery, "max_combo", options.MinCombo, options.MaxCombo)
	addRangeQuery(&boolQuery, "date_submitted", options.MinDateSubmitted, options.MaxDateSubmitted)
	addRangeQuery(&boolQuery, "date_last_updated", options.MinLastUpdated, options.MaxLastUpdated)

	if !options.Explicit {
		explicitTerm := TermCustom{}
		explicitTerm.Term.Explicit = &Term{
			Value: options.Explicit,
		}
		boolQuery.BoolQuery.Must = append(boolQuery.BoolQuery.Must, explicitTerm)
	}

	validSortFields := map[string]bool{
		"length":               true,
		"difficulty_rating":    true,
		"max_combo":            true,
		"play_count":           true,
		"bpm":                  true,
		"long_note_percentage": true,
		"date_submitted":       true,
		"date_last_updated":    true,
	}

	sort := "date_last_updated"
	sortOrder := "desc"

	if options.SortBy != "" && validSortFields[options.SortBy] {
		sort = options.SortBy
	}

	if options.SortOrder == "asc" || options.SortOrder == "desc" {
		sortOrder = options.SortOrder
	}

	if options.IsClanRanked {
		clanRankedTerm := TermCustom{}
		clanRankedTerm.Term.IsClanRanked = &Term{
			Value: options.IsClanRanked,
		}
		boolQuery.BoolQuery.Must = append(boolQuery.BoolQuery.Must, clanRankedTerm)

		sort = "date_clan_ranked"
	}

	query := Query{
		Size: options.Limit,
		From: options.Page * options.Limit,
		Collapse: Collapse{
			Field: "mapset_id",
			InnerHits: InnerHits{
				Name: "most_relevant",
				Size: 50,
				Sort: []map[string]SortOrder{
					{"difficulty_rating": {Order: "asc"}},
				},
			},
		},
		Query: boolQuery,
		Sort: []map[string]SortOrder{
			{"_score": {Order: "desc"}},
			{sort: {Order: sortOrder}},
		},
		Aggs: map[string]interface{}{
			"distinct_mapset_ids": map[string]interface{}{
				"cardinality": map[string]interface{}{
					"field": "mapset_id",
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)

	if err != nil {
		return nil, 0, errors.New(fmt.Sprintf("Error marshaling the query: %s", err))
	}

	resp, err := ElasticSearch.Search(
		ElasticSearch.Search.WithIndex(elasticMapSearchIndex),
		ElasticSearch.Search.WithBody(strings.NewReader(string(queryJSON))),
	)

	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, 0, err
	}

	var hits ElasticHits

	if err := json.Unmarshal(body, &hits); err != nil {
		return nil, 0, err
	}

	var mapsets = make([]*Mapset, 0)

	for _, hit := range hits.Hits.Hits {
		firstHit := hit.InnerHits.MostRelevant.Hits.Hits[0].Source

		mapset := &Mapset{
			Id:                  firstHit.MapsetId,
			PackageMD5:          firstHit.PackageMD5,
			CreatorID:           firstHit.CreatorId,
			CreatorUsername:     firstHit.CreatorUsername,
			Artist:              firstHit.Artist,
			Title:               firstHit.Title,
			Source:              firstHit.Source,
			Tags:                firstHit.Tags,
			Description:         firstHit.Description,
			Maps:                []*MapQua{},
			DateSubmitted:       firstHit.DateSubmitted,
			DateSubmittedJSON:   time.UnixMilli(firstHit.DateSubmitted),
			DateLastUpdated:     firstHit.DateLastUpdated,
			DateLastUpdatedJSON: time.UnixMilli(firstHit.DateLastUpdated),
			IsVisible:           true,
		}

		mapsets = append(mapsets, mapset)

		for _, mapQua := range hit.InnerHits.MostRelevant.Hits.Hits {
			mapset.Maps = append(mapset.Maps, mapQua.Source.MapQua)
		}
	}

	return mapsets, hits.Aggregations.Distinct.Total, nil
}
