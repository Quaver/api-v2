package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRankingQueueComments Returns all of the comments for a mapset in the ranking queue
// Endpoint: GET /v2/ranking/queue/:id/comments
func GetRankingQueueComments(c *gin.Context) *APIError {
	c.JSON(http.StatusOK, gin.H{})
	return nil
}
