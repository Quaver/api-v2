package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
)

func TestGetMapsetsSearchRejectsInvalidAdvancedSearch(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodGet, "/v2/mapset/search?advanced_search=d%3C10%20d%3E20", nil)

	apiErr := GetMapsetsSearch(context)
	if apiErr == nil || apiErr.Status != http.StatusBadRequest {
		t.Fatalf("API error = %#v, want bad request", apiErr)
	}
}

func TestGetUserMapsetRankedStatuses(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		status []enums.RankedStatus
	}{
		{name: "default", query: "", status: []enums.RankedStatus{enums.RankedStatusUnranked, enums.RankedStatusRanked}},
		{name: "ranked", query: "status=2", status: []enums.RankedStatus{enums.RankedStatusRanked}},
		{name: "not submitted", query: "status=0", status: []enums.RankedStatus{enums.RankedStatusNotSubmitted}},
		{name: "malformed", query: "status=invalid", status: []enums.RankedStatus{enums.RankedStatusRanked}},
		{name: "empty", query: "status=", status: []enums.RankedStatus{enums.RankedStatusRanked}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest("GET", "/v2/user/123/mapsets?"+test.query, nil)

			if got := getUserMapsetRankedStatuses(context); len(got) != len(test.status) ||
				len(got) > 0 && (got[0] != test.status[0] || len(got) > 1 && got[1] != test.status[1]) {
				t.Fatalf("statuses = %v, want %v", got, test.status)
			}
		})
	}
}
