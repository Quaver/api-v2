package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// GetUserScoresForClanScore Retrieves all individual scores that make up a clan score
// Endpoint: /v2/clan/scores/:id
func GetUserScoresForClanScore(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	clanScore, err := db.GetClanScoreById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving clan score from db", err)
	}

	if clanScore == nil {
		return APIErrorNotFound("Clan score")
	}

	scores, err := db.GetClanPlayerScoresOnMap(clanScore.MapMD5, clanScore.ClanId, true)

	if err != nil {
		return APIErrorServerError("Error retrieving player clan scores from db", err)
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
	return nil
}
