package db

import (
	"github.com/Quaver/api2/config"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
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
