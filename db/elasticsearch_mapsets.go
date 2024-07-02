package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Quaver/api2/tasks"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"time"
)

type ElasticMapsetSearchOptions struct {
	Limit int // The amount of mapsets to retrieve
}

// NewMapsetSearchOptions Returns a new search options object with default values
func NewMapsetSearchOptions() *ElasticMapsetSearchOptions {
	return &ElasticMapsetSearchOptions{
		Limit: 50,
	}
}

// IndexElasticSearchMapset Indexes an individual mapset in elastic
func IndexElasticSearchMapset(mapset Mapset) error {
	if err := DeleteElasticSearchMapset(mapset.Id); err != nil {
		return err
	}

	mapset.User = nil
	data, _ := json.Marshal(&mapset)

	resp, err := ElasticSearch.Create(elasticMapsetIndex,
		fmt.Sprintf("%v", mapset.Id), bytes.NewReader(data))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// UpdateElasticSearchMapset Updates individual mapset in elastic
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
	resp, err := ElasticSearch.Delete(elasticMapsetIndex, fmt.Sprintf("%v", id))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// IndexAllElasticSearchMapsets Indexes all mapsets in the DB in ElasticSearch
func IndexAllElasticSearchMapsets(deletePrevious bool, workers int) error {
	if deletePrevious {
		if err := DeleteElasticIndices(elasticMapsetIndex); err != nil {
			return err
		}
	}

	pool := tasks.NewWorkerPool(workers)

	// Takes in a mapset and returns a bool if it was indexed successfully.
	pool.Start(func(input ...interface{}) (interface{}, error) {
		mapset := input[0].(*Mapset)

		attempts := 0
		const maxAttempts int = 10

		for attempts < maxAttempts {
			err := IndexElasticSearchMapset(*mapset)

			if err != nil {
				attempts++

				logrus.Errorf("Failed to index mapset #%v. Retrying in 10 seconds...", mapset.Id)
				time.Sleep(time.Second * 10)
				continue
			}

			return true, nil
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
	query := map[string]interface{}{
		"size": options.Limit,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	jsonData, err := json.Marshal(query)

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
