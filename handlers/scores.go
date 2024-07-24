package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type userScoreParams struct {
	Id   int
	User *db.User
	Mode enums.GameMode
	Page int
}

// Function that parses and returns a struct containing recurring data to query user scores.
// Example: user best, recent, and first place scores.
func parseUserScoreParams(c *gin.Context) (*userScoreParams, *APIError) {
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

	user, apiErr := getUserById(id, canAuthedUserViewBannedUsers(c))

	if apiErr != nil {
		return nil, apiErr
	}

	return &userScoreParams{
		Id:   id,
		User: user,
		Mode: enums.GameMode(mode),
		Page: page,
	}, nil
}

// GetUserBestScoresForMode Gets the user's best scores for a given game mode
// Endpoint: /v2/user/:id/scores/:mode/best
func GetUserBestScoresForMode(c *gin.Context) *APIError {
	query, apiErr := parseUserScoreParams(c)

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

	query, apiErr := parseUserScoreParams(c)

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

// GetUserFirstPlaceScoresForMode Gets a user's first place scores for a given game mode
// Endpoint: /v2/user/:id/scores/:mode/firstplace
func GetUserFirstPlaceScoresForMode(c *gin.Context) *APIError {
	query, apiErr := parseUserScoreParams(c)

	if apiErr != nil {
		return apiErr
	}

	scores, err := db.GetUserFirstPlaceScoresForMode(query.Id, query.Mode, 50, query.Page)

	if err != nil {
		return APIErrorServerError("Error retrieving scores from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetUserGradesForMode Gets a user's scores with a particular grade
// Endpoint: GET /v2/user/:id/scores/:mode/grade/:grade
func GetUserGradesForMode(c *gin.Context) *APIError {
	query, apiErr := parseUserScoreParams(c)

	if apiErr != nil {
		return apiErr
	}

	scores, err := db.GetUserGradeScoresForMode(query.Id, query.Mode, c.Param("grade"), 50, query.Page)

	if err != nil {
		return APIErrorServerError("Error retrieving scores from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetPinnedScoresForMode Gets a user's pinned scores for a given game mode
// Endpoint: GET /v2/user/:id/scores/:mode/pinned
func GetPinnedScoresForMode(c *gin.Context) *APIError {
	query, apiErr := parseUserScoreParams(c)

	if apiErr != nil {
		return apiErr
	}

	if !enums.HasUserGroup(query.User.UserGroups, enums.UserGroupDonator) {
		c.JSON(http.StatusOK, gin.H{"scores": []*db.PinnedScore{}})
		return nil
	}

	scores, err := db.GetUserPinnedScores(query.Id, query.Mode)

	if err != nil {
		return APIErrorServerError("Error retrieving user pinned scores", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}
