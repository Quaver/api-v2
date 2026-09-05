package db

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Quaver/api2/enums"
	"github.com/elastic/go-elasticsearch/v8"
)

func TestParseAdvancedSearch(t *testing.T) {
	parsed, err := parseAdvancedSearch("Camellia t=sv t=mixed_ln t=sv k=4 k=7 k=4 s=r s=u s=c d>2 d<=10 b>=120 l=90 ln<30 pc>=100 lu<=1757001600000")
	if err != nil {
		t.Fatal(err)
	}

	if parsed.text != "Camellia" {
		t.Fatalf("text = %q, want Camellia", parsed.text)
	}

	if got, want := parsed.modes, []enums.GameMode{enums.GameModeKeys4, enums.GameModeKeys7}; !sameGameModes(got, want) {
		t.Fatalf("modes = %v, want %v", got, want)
	}

	if got, want := parsed.statuses, []enums.RankedStatus{enums.RankedStatusRanked, enums.RankedStatusUnranked}; !sameRankedStatuses(got, want) {
		t.Fatalf("statuses = %v, want %v", got, want)
	}
	if got, want := strings.Join(parsed.tags, ","), "sv,mixed ln"; got != want {
		t.Fatalf("tags = %q, want %q", got, want)
	}
	if !parsed.clanRanked {
		t.Fatal("clanRanked = false, want true")
	}

	assertRange(t, parsed.ranges["difficulty_rating"], false, 2, false, 10, true)
	assertRange(t, parsed.ranges["bpm"], false, 120, true, 0, false)
	assertRange(t, parsed.ranges["length"], false, 90, true, 90, true)
	assertRange(t, parsed.ranges["long_note_percentage"], false, 0, false, 30, false)
	assertRange(t, parsed.ranges["play_count"], true, 100, true, 0, false)
	assertRange(t, parsed.ranges["date_last_updated"], true, 0, false, 1757001600000, true)
}

func TestGameModeFromKeyCount(t *testing.T) {
	want := []enums.GameMode{
		enums.GameModeKeys1, enums.GameModeKeys2, enums.GameModeKeys3, enums.GameModeKeys4, enums.GameModeKeys5,
		enums.GameModeKeys6, enums.GameModeKeys7, enums.GameModeKeys8, enums.GameModeKeys9, enums.GameModeKeys10,
	}

	for keyCount, mode := range want {
		input := "10"
		if keyCount < 9 {
			input = string(rune('1' + keyCount))
		}
		got, err := gameModeFromKeyCount(input)
		if err != nil || got != mode {
			t.Fatalf("key count %d = %v, %v; want %v", keyCount+1, got, err, mode)
		}
	}
}

func TestParseAdvancedSearchRejectsInvalidFilters(t *testing.T) {
	for _, expression := range []string{
		"d<10 d>20",
		"d<=10 d>10",
		"z=10",
		"d=invalid",
		"pc=1.5",
		"lu=1.5",
		"k=11",
		"s=n",
		"t=unknown",
		"t>sv",
		"d<",
	} {
		t.Run(expression, func(t *testing.T) {
			if _, err := parseAdvancedSearch(expression); err == nil {
				t.Fatal("expected parser error")
			}
		})
	}
}

func TestApplyAdvancedSearchOverridesMatchingLegacyFilters(t *testing.T) {
	options := NewElasticMapsetSearchOptions()
	options.Search = "legacy search"
	options.MinDifficultyRating = 1
	options.MaxDifficultyRating = 20
	options.MinBPM = 120
	options.AdvancedSearch = "Camellia d>2 d<10"

	if err := options.ApplyAdvancedSearch(); err != nil {
		t.Fatal(err)
	}

	if options.Search != "Camellia" {
		t.Fatalf("search = %q, want Camellia", options.Search)
	}
	assertRange(t, options.advancedRanges["difficulty_rating"], false, 2, false, 10, false)
	if _, exists := options.advancedRanges["bpm"]; exists {
		t.Fatal("advanced BPM range unexpectedly present")
	}
}

func TestBindAndValidateDoesNotApplyAdvancedSearch(t *testing.T) {
	options := NewUserElasticMapsetSearchOptions(123)
	options.AdvancedSearch = "d<10"
	options.BindAndValidate()

	if options.Search != "" {
		t.Fatalf("search = %q, want empty", options.Search)
	}
	if options.advancedRanges != nil {
		t.Fatalf("advanced ranges = %#v, want nil", options.advancedRanges)
	}
}

func TestSearchElasticMapsetsAdvancedSearchQuery(t *testing.T) {
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
		_, _ = writer.Write([]byte(`{"hits":{"hits":[]},"aggregations":{"distinct_mapset_ids":{"value":0}}}`))
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
	options.Search = "legacy"
	options.MinDifficultyRating = 1
	options.MaxDifficultyRating = 20
	options.MinBPM = 120
	options.AdvancedSearch = "Camellia t=sv t=mixed_ln k=4 k=7 s=r s=u s=c d>2 d<=10"
	options.BindAndValidate()
	if err := options.ApplyAdvancedSearch(); err != nil {
		t.Fatal(err)
	}

	if _, _, err := SearchElasticMapsets(options); err != nil {
		t.Fatal(err)
	}

	serialized, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatal(err)
	}
	query := string(serialized)
	for _, want := range []string{
		`"query":"camellia"`,
		`"difficulty_rating":{"gt":2,"lte":10}`,
		`"bpm":{"gte":120`,
		`"game_mode":{"value":1`,
		`"game_mode":{"value":2`,
		`"ranked_status":{"value":2`,
		`"ranked_status":{"value":1`,
		`"is_clan_ranked":{"value":true`,
		`"match":{"tags":{"query":"sv"`,
		`"match":{"tags":{"query":"mixed ln"`,
		`"minimum_should_match":1`,
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query does not contain %s: %s", want, query)
		}
	}

	if strings.Contains(query, `"difficulty_rating":{"gte":1`) || strings.Contains(query, `"difficulty_rating":{"lte":20`) {
		t.Fatalf("legacy difficulty range was not overridden: %s", query)
	}
}

func assertRange(t *testing.T, got advancedRangeConstraint, integer bool, lower float64, lowerInclusive bool, upper float64, upperInclusive bool) {
	t.Helper()
	if got.integer != integer {
		t.Fatalf("integer = %v, want %v", got.integer, integer)
	}
	if lower == 0 && !lowerInclusive {
		if got.lower != nil {
			t.Fatalf("lower = %#v, want nil", got.lower)
		}
	} else if got.lower == nil || got.lower.inclusive != lowerInclusive || rangeValue(*got.lower, integer) != lower {
		t.Fatalf("lower = %#v, want value %v inclusive %v", got.lower, lower, lowerInclusive)
	}
	if upper == 0 && !upperInclusive {
		if got.upper != nil {
			t.Fatalf("upper = %#v, want nil", got.upper)
		}
	} else if got.upper == nil || got.upper.inclusive != upperInclusive || rangeValue(*got.upper, integer) != upper {
		t.Fatalf("upper = %#v, want value %v inclusive %v", got.upper, upper, upperInclusive)
	}
}

func rangeValue(bound advancedRangeBound, integer bool) float64 {
	if integer {
		return float64(bound.intValue)
	}
	return bound.floatValue
}

func sameGameModes(left, right []enums.GameMode) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameRankedStatuses(left, right []enums.RankedStatus) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
