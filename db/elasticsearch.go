package db

import (
	"github.com/Quaver/api2/config"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

var ElasticSearch *elasticsearch.Client

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
	logrus.Info("Successfully initialized ElasticSearch")
}
