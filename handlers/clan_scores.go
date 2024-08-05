package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetClanScoresForMode Retrieves clan scores for a given mode
// Endpoint: GET /v2/clan/:id/scores/:mode
func GetClanScoresForMode(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mode, err := strconv.Atoi(c.Param("mode"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	scores, err := db.GetClanScoresForModeFull(id, enums.GameMode(mode), page)

	if err != nil {
		return APIErrorServerError("Error retrieving clan scores from db", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}
