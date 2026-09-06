package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetUserActivity Gets the user's recent activity
// Endpoint: GET /v2/user/:id/activity
func GetUserActivity(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	limit := 50
	requestedLimit, err := strconv.Atoi(c.Query("limit"))

	if err == nil && requestedLimit > 0 {
		limit = min(requestedLimit, limit)
	}

	if _, apiErr := getUserById(id, canAuthedUserViewBannedUsers(c)); apiErr != nil {
		return apiErr
	}

	activities, err := db.GetRecentUserActivity(id, limit, page)

	if err != nil {
		return APIErrorServerError("Error getting user activities", err)
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
	return nil
}
