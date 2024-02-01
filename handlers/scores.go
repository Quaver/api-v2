package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type userScoreQuery struct {
	Id   int
	Mode enums.GameMode
	Page int
}

// Function that parses and returns a struct containing recurring data to query user scores.
// Example: user best, recent, and first place scores.
func parseUserScoreQuery(c *gin.Context) (*userScoreQuery, *APIError) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return nil, APIErrorBadRequest("Invalid id")
	}

	mode, err := strconv.Atoi(c.Param("mode"))

	if err != nil {
		return nil, APIErrorBadRequest("Invalid mode")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	return &userScoreQuery{
		Id:   id,
		Mode: enums.GameMode(mode),
		Page: page,
	}, nil
}

// GetUserBestScoresForMode Gets the user's best scores for a given game mode
// Endpoint: /v2/user/:id/scores/:mode/best
func GetUserBestScoresForMode(c *gin.Context) *APIError {
	query, apiErr := parseUserScoreQuery(c)

	if apiErr != nil {
		return apiErr
	}

	scores, err := db.GetUserBestScoresForMode(query.Id, query.Mode, 50, query.Page)

	if err != nil {
		return APIErrorServerError("Error retrieving scores from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetUserRecentScoresForMode Gets the user's recent scores for a given game mode
// Endpoint: /v2/user/:id/scores/:mode/recent
func GetUserRecentScoresForMode(c *gin.Context) *APIError {
	isDonator := false

	if user := getAuthedUser(c); user != nil && enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		isDonator = true
	}

	query, apiErr := parseUserScoreQuery(c)

	if apiErr != nil {
		return apiErr
	}

	scores, err := db.GetUserRecentScoresForMode(query.Id, query.Mode, isDonator, 50, query.Page)

	if err != nil {
		return APIErrorServerError("Error retrieving scores from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}
