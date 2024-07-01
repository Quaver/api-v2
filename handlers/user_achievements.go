package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetUserAchievements Returns a list of achievements that a user has locked/unlocked
// Endpoint: GET /v2/user/:id/achievements
func GetUserAchievements(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	if _, apiErr := getUserById(id, canAuthedUserViewBannedUsers(c)); apiErr != nil {
		return apiErr
	}

	achievements, err := db.GetUserAchievements(id)

	if err != nil {
		return APIErrorServerError("Error getting user achievements", err)
	}

	c.JSON(http.StatusOK, gin.H{"achievements": achievements})
	return nil
}
