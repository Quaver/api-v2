package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetUserMostPlayedMaps Gets a user's most played maps
// Endpoint: /v2/user/:id/mostplayed
func GetUserMostPlayedMaps(c *gin.Context) *APIError {
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

	maps, err := db.GetUserMostPlayedMaps(id, 10, page)

	if err != nil {
		return APIErrorServerError("Error retrieving most played maps in db", err)
	}

	c.JSON(http.StatusOK, gin.H{"maps": maps})
	return nil
}

// GetWeeklyMostPlayedMapsets Gets the most played mapsets of the week
// Endpoint: GET /v2/server/stats/mostplayed
func GetWeeklyMostPlayedMapsets(c *gin.Context) *APIError {
	mapsets, err := db.GetWeeklyMostPlayedMapsets()

	if err != nil {
		return APIErrorServerError("Error retrieving weekly most played mapsets in db", err)
	}

	c.JSON(http.StatusOK, gin.H{"mapsets": mapsets})
	return nil
}
