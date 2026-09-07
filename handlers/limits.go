package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultUserScoreLimit    = 50
	defaultClanScoreLimit    = 50
	defaultMostPlayedLimit   = 10
	defaultUserActivityLimit = 50
	defaultUserPlaylistLimit = 50
)

// getQueryLimit returns a requested limit bounded by the endpoint's default limit.
// Invalid values use the endpoint's default limit.
func getQueryLimit(c *gin.Context, defaultLimit int) int {
	limit, err := strconv.Atoi(c.Query("limit"))

	if err != nil || limit < 1 {
		return defaultLimit
	}

	if limit > defaultLimit {
		return defaultLimit
	}

	return limit
}
