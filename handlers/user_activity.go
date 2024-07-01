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

	if _, apiErr := getUserById(id); apiErr != nil {
		return apiErr
	}

	activities, err := db.GetRecentUserActivity(id, 50, page)

	if err != nil {
		return APIErrorServerError("Error getting user activities", err)
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
	return nil
}
