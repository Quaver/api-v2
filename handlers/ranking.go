package handlers

import (
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetRankingQueue Retrieves the ranking queue for a given mode
// Endpoint: /v2/ranking/queue/:mode
func GetRankingQueue(c *gin.Context) *APIError {
	mode, err := strconv.Atoi(c.Param("mode"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mode")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	rankingQueue, err := db.GetRankingQueue(enums.GameMode(mode), 20, page)

	if err != nil {
		return APIErrorServerError("Error retrieving ranking queue", err)
	}

	c.JSON(http.StatusOK, gin.H{"ranking_queue": rankingQueue})
	return nil
}

// GetRankingQueueConfig Returns the vote/denial configuration for the ranking queue
func GetRankingQueueConfig(c *gin.Context) *APIError {
	c.JSON(http.StatusOK, gin.H{
		"votes_required":          config.Instance.RankingQueue.VotesRequired,
		"denials_required":        config.Instance.RankingQueue.DenialsRequired,
		"mapset_uploads_required": config.Instance.RankingQueue.MapsetUploadsRequired,
	})

	return nil
}
