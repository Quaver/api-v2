package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetUserRankStatisticsForMode Gets a user's rank statistics for a given game mode
// Endpoint: GET /v2/user/:id/statistics/:mode/rank
func GetUserRankStatisticsForMode(c *gin.Context) *APIError {
	query, apiErr := parseUserScoreQuery(c)

	if apiErr != nil {
		return apiErr
	}

	ranks, err := db.GetUserRankStatisticsForMode(query.Id, query.Mode)

	if err != nil {
		return APIErrorServerError("Error getting rank statistics", err)
	}

	c.JSON(http.StatusOK, gin.H{"ranks": ranks})
	return nil
}
