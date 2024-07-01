package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetUserBadges Gets a user's profile badges
// Endpoint: GET /v2/user/:id/badges
func GetUserBadges(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	if _, apiErr := getUserById(id, canAuthedUserViewBannedUsers(c)); apiErr != nil {
		return apiErr
	}

	badges, err := db.GetUserBadges(id)

	if err != nil {
		return APIErrorServerError("Error getting user badges", err)
	}

	c.JSON(http.StatusOK, gin.H{"badges": badges})
	return nil
}
