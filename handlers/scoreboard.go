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
	dbMap, err := db.GetMapByMD5(c.Param("md5"))

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		c.JSON(404, gin.H{"message": "The map you have provided was not found."})
		return nil
	default:
		return APIErrorServerError("Error retrieving map from database", err)
	}

	limit := 50

	user := getAuthedUser(c)

	if user != nil {
		if enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
			limit = 100
		}
	}

	// Only allow donators to access non-ranked map scoreboards
	if dbMap.RankedStatus != enums.RankedStatusRanked {
		if user == nil || !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
			return APIErrorForbidden("You must be a donator to access this scoreboard.")
		}
	}

	scores, err := db.GetGlobalScoresForMap(dbMap.MD5, limit, 0)

	if err != nil {
		return APIErrorServerError("Error retrieving global scoreboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}
