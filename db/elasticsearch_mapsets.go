package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type ElasticMapsetSearchOptions struct {
	Search string               `form:"search" json:"search"`
	Status []enums.RankedStatus `form:"status" json:"status"`
	Mode   []enums.GameMode     `form:"mode" json:"mode"`
	Page   int                  `form:"page" json:"page"`
	Limit  int                  `form:"limit" json:"limit"`

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
}

type ElasticMap struct {
	*MapQua
	PackageMD5      string `json:"package_md5"`
	DateSubmitted   int64  `json:"date_submitted"`
	DateLastUpdated int64  `json:"date_last_updated"`
}

type ElasticAggregation struct {
	Aggregations struct {
		ByMapsetID struct {
			Buckets []struct {
				GroupedHits struct {
					Hits struct {
						Hits []struct {
							ID     string     `json:"_id"`
							Source ElasticMap `json:"_source"`
						} `json:"hits"`
					} `json:"hits"`
				} `json:"grouped_hits"`
				Key      int `json:"key,omitempty"`
				DocCount int `json:"doc_count,omitempty"`
			} `json:"buckets"`
		} `json:"by_mapset_id"`
	} `json:"aggregations"`
}

// NewMapsetSearchOptions Returns a new search options object with default values
func NewMapsetSearchOptions() *ElasticMapsetSearchOptions {
	return &ElasticMapsetSearchOptions{
		Search: "*",
		Limit:  50,
	}
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
		}

		data, err := json.Marshal(&elasticMap)

		if err != nil {
			return err
		}

		resp, err := ElasticSearch.Update(elasticMapsetIndex, fmt.Sprintf("%v", mapQua.Id), bytes.NewReader(data))

		if err != nil {
			return err
		}

		defer resp.Body.Close()
	}

	return nil
}

// DeleteElasticSearchMapset Deletes an individual mapset in elastic
func DeleteElasticSearchMapset(id int) error {
	queryMap := TermQuery{
		Query: Term{
			Term: Field{
				MapsetID: id,
			},
		},
	}

	queryJSON, err := json.Marshal(queryMap)

	if err != nil {
		log.Fatalf("Error marshalling query to JSON: %s", err)
	}

	resp, err := ElasticSearch.DeleteByQuery([]string{elasticMapSearchIndex}, strings.NewReader(string(queryJSON)))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// IndexAllElasticSearchMapsets Indexes all mapsets in the DB in ElasticSearch
func IndexAllElasticSearchMapsets(deletePrevious bool, workers int) error {
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
		Index:         elasticMapSearchIndex, // The default index name
		Client:        ElasticSearch,         // The Elasticsearch client
		NumWorkers:    5,                     // The number of worker goroutines
		FlushBytes:    int(5e+6),             // The flush threshold in bytes
		FlushInterval: 30 * time.Second,      // The periodic flush interval
	})

	if err != nil {
		log.Fatalf("Error creating the indexer: %s", err)
	}

	var countSuccessful uint64

	// Put all mapsets into the task queue
	for _, mapset := range mapsets {
		for _, mapQua := range mapset.Maps {
			elasticMap := ElasticMap{
				MapQua:          mapQua,
				PackageMD5:      mapset.PackageMD5,
				DateSubmitted:   mapset.DateSubmitted,
				DateLastUpdated: mapset.DateLastUpdated,
			}

			data, err := json.Marshal(&elasticMap)

			if err != nil {
				return err
			}

			err = bi.Add(
				context.Background(),
				esutil.BulkIndexerItem{
					Action: "index",

					// DocumentID is the (optional) document ID
					DocumentID: strconv.Itoa(mapQua.Id),

					// Body is an `io.Reader` with the payload
					Body: bytes.NewReader(data),

					// OnSuccess is called for each successful operation
					OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
						atomic.AddUint64(&countSuccessful, 1)
					},

					// OnFailure is called for each failed operation
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
				log.Fatalf("Unexpected error: %s", err)
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
func SearchElasticMapsets(options *ElasticMapsetSearchOptions) ([]*Mapset, error) {
	query := Query{
		Size: 0,
		From: 1, // Current page
		Aggs: Aggs{
			ByMapsetID: ByMapsetID{
				Terms: Terms{
					Field: "mapset_id",
					Size:  options.Limit, // Results per page
					Order: map[string]string{"_key": "desc"},
				},
				Aggs: GroupedHits{
					GroupedHits: TopHitsAgg{
						TopHits: TopHits{
							Source: map[string]interface{}{},
							Size:   50, // How many maps to group by
							Sort: []map[string]SortOrder{
								{"date_last_updated": {Order: "desc"}},
							},
						},
					},
				},
			},
		},
		Query: BoolQuery{
			Bool: MustBool{
				Must: MustQuery{
					Bool: ShouldBool{
						Should: []QueryString{},
					},
				},
			},
		},
		Sort: []map[string]SortOrder{
			{"date_last_updated": {Order: "desc"}},
		},
	}

	if options.Search != "" {
		query.Query.Bool.Must.Bool.Should = append(query.Query.Bool.Must.Bool.Should, QueryString{
			QueryString: QueryStringParams{
				Query:           options.Search,
				Fields:          []string{"title", "artist"},
				DefaultOperator: "OR",
				Boost:           1.0,
			},
		})
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.Fatalf("Error marshaling the query: %s", err)
	}

	fmt.Printf("%v\n", string(queryJSON))

	resp, err := ElasticSearch.Search(
		ElasticSearch.Search.WithIndex(elasticMapSearchIndex),
		ElasticSearch.Search.WithBody(strings.NewReader(string(queryJSON))),
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var aggregation ElasticAggregation

	if err := json.Unmarshal(body, &aggregation); err != nil {
		return nil, err
	}

	var mapsets []*Mapset

	for _, bucket := range aggregation.Aggregations.ByMapsetID.Buckets {
		if len(bucket.GroupedHits.Hits.Hits) > 0 {
			firstHit := bucket.GroupedHits.Hits.Hits[0].Source

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
				Maps:                []*MapQua{firstHit.MapQua},
				DateSubmitted:       firstHit.DateSubmitted,
				DateSubmittedJSON:   time.UnixMilli(firstHit.DateSubmitted),
				DateLastUpdated:     firstHit.DateLastUpdated,
				DateLastUpdatedJSON: time.UnixMilli(firstHit.DateLastUpdated),
				IsVisible:           true,
			}

			for _, hit := range bucket.GroupedHits.Hits.Hits[1:] {
				mapset.Maps = append(mapset.Maps, hit.Source.MapQua)
			}

			mapsets = append(mapsets, mapset)
		}
	}

	return mapsets, nil
}
