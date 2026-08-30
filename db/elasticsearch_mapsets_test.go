package db

import "testing"

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
