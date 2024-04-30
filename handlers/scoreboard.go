package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
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

	if !checkUserHasScoreboardAccess(dbMap, user) {
		return APIErrorForbidden("You must be a donator to access this scoreboard.")
	}

	scores, err := db.GetGlobalScoresForMap(dbMap.MD5, limit, 0)

	if err != nil {
		return APIErrorServerError("Error retrieving global scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}

// Retrieves the map to be used for the scoreboard. Returns an api error in the event it can't.
func getScoreboardMap(c *gin.Context) (*db.MapQua, *APIError) {
	dbMap, err := db.GetMapByMD5(c.Param("md5"))

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		c.JSON(404, gin.H{"message": "The map you have provided was not found."})
		return nil, nil
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
func checkUserHasScoreboardAccess(dbMap *db.MapQua, user *db.User) bool {
	if dbMap.RankedStatus != enums.RankedStatusRanked {
		if user == nil || !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
			return false
		}
	}

	return true
}