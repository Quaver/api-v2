package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// GetGlobalScoresForMap Retrieves the global scoreboard for a given map.
// Endpoint: GET: /v2/scores/:md5/global
func GetGlobalScoresForMap(c *gin.Context) *APIError {
	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	user := getAuthedUser(c)
	limit := getScoreboardScoreLimit(user)

	if !hasDonatorScoreboardAccess(dbMap, user) {
		return APIErrorForbidden("You must be a donator to access this scoreboard.")
	}

	scores, err := db.GetGlobalScoresForMap(dbMap.MD5, limit, 0)

	if err != nil {
		return APIErrorServerError("Error retrieving global scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetCountryScoresForMap Retrieves country scores for a given map
// Endpoint: GET v2/scores/:md5/country/:country
func GetCountryScoresForMap(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You must be a donator to access this scoreboard.")
	}

	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	limit := getScoreboardScoreLimit(user)
	scores, err := db.GetCountryScoresForMap(dbMap.MD5, c.Param("country"), limit, 0)

	if err != nil {
		return APIErrorServerError("Error retrieving country scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetModifierScoresForMap Retrieves modifier scores for a given map
// Endpoint: GET v2/scores/:md5/mods/:mods
func GetModifierScoresForMap(c *gin.Context) *APIError {
	mods, err := strconv.ParseInt(c.Param("mods"), 10, 64)

	if err != nil {
		return APIErrorBadRequest("You must provide a valid modifier value.")
	}

	// No mods active, so it's the same as the global leaderboard.
	if mods == 0 {
		return GetGlobalScoresForMap(c)
	}

	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	user := getAuthedUser(c)
	limit := getScoreboardScoreLimit(user)

	if !hasDonatorScoreboardAccess(dbMap, user) {
		return APIErrorForbidden("You must be a donator to access this scoreboard.")
	}

	scores, err := db.GetModifierScoresForMap(dbMap.MD5, mods, limit, 0)

	if err != nil {
		return APIErrorServerError("Error retrieving modifier scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetRateScoresForMap Retrieves rate scores for a given map
// Endpoint: GET v2/scores/:md5/rate/:mods
func GetRateScoresForMap(c *gin.Context) *APIError {
	mods, err := strconv.ParseInt(c.Param("mods"), 10, 64)

	if err != nil {
		return APIErrorBadRequest("You must provide a valid modifier value.")
	}

	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	user := getAuthedUser(c)
	limit := getScoreboardScoreLimit(user)

	if !hasDonatorScoreboardAccess(dbMap, user) {
		return APIErrorForbidden("You must be a donator to access this scoreboard.")
	}

	scores, err := db.GetRateScoresForMap(dbMap.MD5, mods, limit, 0)

	if err != nil {
		return APIErrorServerError("Error retrieving rate scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetAllScoresForMap Retrieves all scores for a given map
// Endpoint: GET v2/scores/:md5/all
func GetAllScoresForMap(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You must be a donator to access this scoreboard.")
	}

	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	scores, err := db.GetAllScoresForMap(dbMap.MD5, getScoreboardScoreLimit(user), 0)

	if err != nil {
		return APIErrorServerError("Error retrieving all scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetFriendScoresForMap Retrieves the friends scoreboard for a given map.
// Endpoint: GET: /v2/scores/:md5/friends
func GetFriendScoresForMap(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	friends, err := db.GetUserFriends(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving user friends", err)
	}

	limit := getScoreboardScoreLimit(user)

	if !hasDonatorScoreboardAccess(dbMap, user) {
		return APIErrorForbidden("You must be a donator to access this scoreboard.")
	}

	scores, err := db.GetFriendScoresForMap(dbMap.MD5, user.Id, friends, limit, 0)

	if err != nil {
		return APIErrorServerError("Error retrieving global scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// GetUserPersonalBestScoreGlobal Retrieves the personal best score on a map for a user
// Endpoint: GET /v2/scores/:md5/:user_id/global
func GetUserPersonalBestScoreGlobal(c *gin.Context) *APIError {
	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	userId, err := strconv.Atoi(c.Param("user_id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid user id")
	}

	score, err := db.GetUserPersonalBestScoreGlobal(userId, dbMap.MD5)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving personal best global score from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"score": score})
	return nil
}

// GetUserPersonalBestScoreAll Retrieves the personal best (ALL lb) score on a map for a user
// Endpoint: GET /v2/scores/:md5/:user_id/all
func GetUserPersonalBestScoreAll(c *gin.Context) *APIError {
	dbMap, apiErr := getScoreboardMap(c)

	if apiErr != nil {
		return apiErr
	}

	userId, err := strconv.Atoi(c.Param("user_id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid user id")
	}

	score, err := db.GetUserPersonalBestScoreAll(userId, dbMap.MD5)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving personal best (all-time) score from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"score": score})
	return nil
}

// Retrieves the map to be used for the scoreboard. Returns an api error in the event it can't.
func getScoreboardMap(c *gin.Context) (*db.MapQua, *APIError) {
	dbMap, err := db.GetMapByMD5(c.Param("md5"))

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return nil, &APIError{
			Status:  404,
			Message: "The map you have provided was not found",
			Error:   nil,
		}
	default:
		return nil, APIErrorServerError("Error retrieving map from database", err)
	}

	return dbMap, nil
}

// Returns the amount of scores the user will be able to view.
func getScoreboardScoreLimit(user *db.User) int {
	limit := 50

	if user != nil {
		if enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
			limit = 100
		}
	}

	return limit
}

// Only donators are able to access leaderboards for non-ranked, so run a check for this.
func hasDonatorScoreboardAccess(dbMap *db.MapQua, user *db.User) bool {
	if dbMap.RankedStatus != enums.RankedStatusRanked {
		if user == nil || !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
			return false
		}
	}

	return true
}
