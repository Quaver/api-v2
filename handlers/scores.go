package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetUserBestScoresForMode Gets the user's best scores for a given game mode
// Endpoint: /v2/user/:id/scores/:mode/best
func GetUserBestScoresForMode(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mode, err := strconv.Atoi(c.Param("mode"))

	if err != nil {
		return APIErrorBadRequest("Invalid mode")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	scores, err := db.GetUserBestScoresForMode(id, enums.GameMode(mode), 50, page)

	if err != nil {
		return APIErrorServerError("Error retrieving scores from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}
