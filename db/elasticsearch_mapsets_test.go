package db

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

func TestGetMapsetPlayCount(t *testing.T) {
	mapset := Mapset{
		Maps: []*MapQua{
			{PlayCount: 12},
			{PlayCount: 345},
			{PlayCount: 6789},
		},
	}

	if got := getMapsetPlayCount(mapset); got != 7146 {
		t.Fatalf("getMapsetPlayCount() = %d, want %d", got, 7146)
	}
}

func TestElasticMapsetSortField(t *testing.T) {
	if got := elasticMapsetSortField("play_count"); got != "mapset_play_count" {
		t.Fatalf("elasticMapsetSortField(play_count) = %q, want %q", got, "mapset_play_count")
	}

	if got := elasticMapsetSortField("date_last_updated"); got != "date_last_updated" {
		t.Fatalf("elasticMapsetSortField(date_last_updated) = %q, want %q", got, "date_last_updated")
	}
}

func TestNewUserElasticMapsetSearchOptions(t *testing.T) {
	options := NewUserElasticMapsetSearchOptions(123)

	if options.CreatorID == nil || *options.CreatorID != 123 {
		t.Fatalf("CreatorID = %v, want 123", options.CreatorID)
	}

	if len(options.RankedStatus) != 2 ||
		options.RankedStatus[0] != 1 || options.RankedStatus[1] != 2 {
		t.Fatalf("RankedStatus = %v, want [1 2]", options.RankedStatus)
	}

	if options.Limit != 50 {
		t.Fatalf("Limit = %d, want 50", options.Limit)
	}

	if !options.Explicit {
		t.Fatal("Explicit = false, want true")
	}
}

func TestElasticMapsetSearchOptionsClampLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{name: "below minimum", limit: 0, want: 1},
		{name: "negative", limit: -10, want: 1},
		{name: "within range", limit: 25, want: 25},
		{name: "above maximum", limit: 51, want: 50},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := NewUserElasticMapsetSearchOptions(123)
			options.Limit = test.limit
			options.BindAndValidate()

			if options.Limit != test.want {
				t.Fatalf("Limit = %d, want %d", options.Limit, test.want)
			}
		})
	}
}

func TestSearchElasticMapsetsUserQuery(t *testing.T) {
	var requestBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &requestBody); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Elastic-Product", "Elasticsearch")
		_, _ = writer.Write([]byte(`{
			"hits": {"hits": [{
				"inner_hits": {"most_relevant": {"hits": {"hits": [{
					"_source": {
						"id": 456,
						"mapset_id": 789,
						"creator_id": 123,
						"creator_username": "user",
						"artist": "Artist",
						"title": "Title",
						"package_md5": "package-md5",
						"ranked_status": 1,
						"date_submitted": 1000,
						"date_last_updated": 2000
					}
				}]}}}
			}]},
			"aggregations": {"distinct_mapset_ids": {"value": 1}}
		}`))
	}))
	defer server.Close()

	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatal(err)
	}

	previousClient := ElasticSearch
	ElasticSearch = client
	t.Cleanup(func() { ElasticSearch = previousClient })

	options := NewUserElasticMapsetSearchOptions(123)
	options.Page = 2
	options.Limit = 10
	options.BindAndValidate()

	mapsets, total, err := SearchElasticMapsets(options)
	if err != nil {
		t.Fatal(err)
	}

	if total != 1 || len(mapsets) != 1 {
		t.Fatalf("search result = total %d, mapsets %d; want total 1, mapsets 1", total, len(mapsets))
	}

	if mapsets[0].Id != 789 || mapsets[0].CreatorID != 123 || mapsets[0].PackageMD5 != "package-md5" ||
		len(mapsets[0].Maps) != 1 || mapsets[0].Maps[0].Id != 456 {
		t.Fatalf("unexpected mapped mapset: %#v", mapsets[0])
	}

	if requestBody["from"] != float64(20) || requestBody["size"] != float64(10) {
		t.Fatalf("pagination = from %v, size %v; want from 20, size 10", requestBody["from"], requestBody["size"])
	}

	query := requestBody["query"].(map[string]interface{})
	boolQuery := query["bool"].(map[string]interface{})
	must := boolQuery["must"].([]interface{})

	creatorTerm := must[1].(map[string]interface{})["term"].(map[string]interface{})["creator_id"].(map[string]interface{})
	if creatorTerm["value"] != float64(123) {
		t.Fatalf("creator_id = %v, want 123", creatorTerm["value"])
	}

	statusQuery := must[2].(map[string]interface{})["bool"].(map[string]interface{})
	statusShould := statusQuery["should"].([]interface{})
	if len(statusShould) != 2 {
		t.Fatalf("status terms = %d, want 2", len(statusShould))
	}

	statuses := make(map[float64]bool)
	for _, clause := range statusShould {
		term := clause.(map[string]interface{})["term"].(map[string]interface{})["ranked_status"].(map[string]interface{})
		statuses[term["value"].(float64)] = true
	}

	if !statuses[1] || !statuses[2] {
		t.Fatalf("status terms = %v, want ranked and unranked", statuses)
	}

	collapse := requestBody["collapse"].(map[string]interface{})
	if collapse["field"] != "mapset_id" {
		t.Fatalf("collapse field = %v, want mapset_id", collapse["field"])
	}
}

func TestSearchElasticMapsetsGlobalQueryHasNoCreatorFilter(t *testing.T) {
	var requestBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &requestBody); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Elastic-Product", "Elasticsearch")
		_, _ = writer.Write([]byte(`{
			"hits": {"hits": []},
			"aggregations": {"distinct_mapset_ids": {"value": 0}}
		}`))
	}))
	defer server.Close()

	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{server.URL}})
	if err != nil {
		t.Fatal(err)
	}

	previousClient := ElasticSearch
	ElasticSearch = client
	t.Cleanup(func() { ElasticSearch = previousClient })

	options := NewElasticMapsetSearchOptions()
	options.Explicit = true
	options.BindAndValidate()

	if _, _, err := SearchElasticMapsets(options); err != nil {
		t.Fatal(err)
	}

	query := requestBody["query"].(map[string]interface{})
	boolQuery := query["bool"].(map[string]interface{})
	for _, clause := range boolQuery["must"].([]interface{}) {
		term, exists := clause.(map[string]interface{})["term"].(map[string]interface{})
		if !exists {
			continue
		}

		if _, exists := term["creator_id"]; exists {
			t.Fatal("global query unexpectedly contains creator_id")
		}
	}
}
