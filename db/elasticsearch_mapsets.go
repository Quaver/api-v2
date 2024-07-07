package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/tasks"
	"github.com/sdqri/effdsl"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"time"
)

type ElasticMapsetSearchOptions struct {
	Search string               `form:"search" json:"search"`
	Status []enums.RankedStatus `form:"status" json:"status"`
	Mode   []enums.GameMode     `form:"mode" json:"mode"`
	Page   uint64               `form:"page" json:"page"`
	Limit  uint64               `form:"limit" json:"limit"`

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
	//MinDateSubmitted    int64   `form:"min_date_submitted" json:"min_date_submitted"`
	//MaxDateSubmitted    int64   `form:"max_date_submitted" json:"max_date_submitted"`
	//MinLastUpdated      int64   `form:"min_last_updated" json:"min_last_updated"`
	//MaxLastUpdated      int64   `form:"max_last_updated" json:"max_last_updated"`
}

// NewMapsetSearchOptions Returns a new search options object with default values
func NewMapsetSearchOptions() *ElasticMapsetSearchOptions {
	return &ElasticMapsetSearchOptions{
		Limit: 50,
	}
}

// IndexElasticSearchMapset Indexes an individual mapset in elastic
func IndexElasticSearchMapset(mapQua MapQua) error {
	//if err := DeleteElasticSearchMapset(mapset.Id); err != nil {
	//	return err
	//}

	data, _ := json.Marshal(&mapQua)

	resp, err := ElasticSearch.Create(elasticMapSearchIndex,
		fmt.Sprintf("%v", mapQua.Id), bytes.NewReader(data))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// UpdateElasticSearchMapset Updates an individual mapset in elastic
func UpdateElasticSearchMapset(mapset Mapset) error {
	mapset.User = nil
	data, _ := json.Marshal(&mapset)

	resp, err := ElasticSearch.Update(elasticMapsetIndex, fmt.Sprintf("%v", mapset.Id), bytes.NewReader(data))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// DeleteElasticSearchMapset Deletes an individual mapset in elastic
func DeleteElasticSearchMapset(id int) error {
	// ToDo - needs to use delete by query API
	//resp, err := ElasticSearch.Delete(elasticMapsetIndex, fmt.Sprintf("%v", id))
	//
	//if err != nil {
	//	return err
	//}
	//
	//defer resp.Body.Close()
	return nil
}

// IndexAllElasticSearchMapsets Indexes all mapsets in the DB in ElasticSearch
func IndexAllElasticSearchMapsets(deletePrevious bool, workers int) error {
	if deletePrevious {
		if err := DeleteElasticIndices(elasticMapSearchIndex); err != nil {
			return err
		}
	}

	pool := tasks.NewWorkerPool(workers)

	// Takes in a mapset and returns a bool if it was indexed successfully.
	pool.Start(func(input ...interface{}) (interface{}, error) {
		mapset := input[0].(*Mapset)

		if len(mapset.Maps) == 0 {
			return false, nil
		}

		attempts := 0
		const maxAttempts int = 10

		for attempts < maxAttempts {
			errored := false

			for _, mapQua := range mapset.Maps {
				err := IndexElasticSearchMapset(*mapQua)

				if err != nil {
					errored = true
					continue
				}

				return true, nil
			}

			if errored {
				attempts++

				logrus.Errorf("Failed to index mapset #%v. Retrying in 10 seconds...", mapset.Id)
				time.Sleep(time.Second * 10)
				continue
			}
		}

		return false, errors.New(fmt.Sprintf("too many failed attempts to index: mapset #%v", mapset.Id))
	})

	mapsets, err := GetAllMapsets()

	if err != nil {
		return err
	}

	// Handle results in goroutine before adding tasks to prevent hanging
	go func() {
		for i := 0; i < len(mapsets); i++ {
			result := pool.GetResult()
			mapset := result.Input.(*Mapset)

			if result.Error != nil {
				logrus.Errorf("Error indexing mapset #%v: %v", mapset.Id, result.Error)
				continue
			}

			logrus.Infof("Successfully indexed mapset #%v", mapset.Id)
		}
	}()

	// Put all mapsets into the task queue
	for _, mapset := range mapsets {
		pool.AddTask(mapset)
	}

	return nil
}

// SearchElasticMapsets Searches ElasticSearch for mapsets
func SearchElasticMapsets(options *ElasticMapsetSearchOptions) ([]*Mapset, error) {
	searchRequest, err := effdsl.Define(
		effdsl.WithQuery(
			effdsl.BoolQuery(
				effdsl.Should(
					effdsl.QueryString(options.Search, effdsl.WithFields("title", "artist")),
					effdsl.QueryString(options.Search, effdsl.WithFields("source")),
					effdsl.QueryString(options.Search, effdsl.WithFields("creator_username")),
					effdsl.QueryString(options.Search, effdsl.WithFields("tags")),
				),
			),
		),
		effdsl.WithPaginate(options.Page, options.Limit),
	)

	jsonData, err := json.Marshal(searchRequest)

	if err != nil {
		return nil, err
	}

	resp, err := ElasticSearch.Search(
		ElasticSearch.Search.WithIndex(elasticMapsetIndex),
		ElasticSearch.Search.WithBody(strings.NewReader(string(jsonData))),
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	data := struct {
		Hits struct {
			Hits []struct {
				Source Mapset `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}{}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var mapsets []*Mapset

	for _, hit := range data.Hits.Hits {
		mapsets = append(mapsets, &hit.Source)
	}

	return mapsets, nil
}
