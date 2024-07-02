package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/tasks"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"time"
)

var ElasticSearch *elasticsearch.Client

const elasticMapsetIndex = "mapsets-v2"

// InitializeElasticSearch Initializes the ElasticSearch client
func InitializeElasticSearch() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			config.Instance.ElasticSearch.Host,
		},
	}

	var err error

	ElasticSearch, err = elasticsearch.NewClient(cfg)

	if err != nil {
		panic(err)
	}

	resp, err := ElasticSearch.Info()

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if err := CreateElasticIndex(elasticMapsetIndex); err != nil {
		panic(err)
	}

	logrus.Info("Successfully initialized ElasticSearch")
}

// CreateElasticIndex Creates a new elastic search index by a given name
func CreateElasticIndex(index string) error {
	resp, err := ElasticSearch.Indices.Create(index)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return err
}

// DeleteElasticIndices Deletes one or many elastic search indices
func DeleteElasticIndices(indices ...string) error {
	resp, err := ElasticSearch.Indices.Delete(indices)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return err
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

// DeleteElasticSearchMapset Deletes an individual mapset in elastic
func DeleteElasticSearchMapset(id int) error {
	resp, err := ElasticSearch.Delete(elasticMapsetIndex, fmt.Sprintf("%v", id))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// SearchElasticMapsets Searches ElasticSearch for mapsets
func SearchElasticMapsets() ([]*Mapset, error) {
	query := `{ "size" : 50, "query": { "match_all": {} } }`

	resp, err := ElasticSearch.Search(
		ElasticSearch.Search.WithIndex(elasticMapsetIndex),
		ElasticSearch.Search.WithBody(strings.NewReader(query)),
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
