package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetRankingQueueComments Returns all of the comments for a mapset in the ranking queue
// Endpoint: GET /v2/ranking/queue/:id/comments
func GetRankingQueueComments(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	comments, err := db.GetRankingQueueComments(id)

	if err != nil {
		return APIErrorServerError("Error getting ranking queue comments", err)
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
	return nil
}
