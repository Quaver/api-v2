package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetClanActivity Retrieves clan activity from the database
// GET /v2/clan/:id/activity?page=0
func GetClanActivity(c *gin.Context) *APIError {
	clanId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	activities, err := db.GetClanActivity(clanId, 50, page)

	if err != nil {
		return APIErrorServerError("Error getting clan activities", err)
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
	return nil
}
