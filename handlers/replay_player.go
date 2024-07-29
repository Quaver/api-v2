package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/files"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetVirtualReplayPlayerOutput Plays the virtual replay player & returns the output
// Endpoint: GET /v2/scores/:id/stats
func GetVirtualReplayPlayerOutput(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	score, err := db.GetScoreById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving score from db", err)
	}

	if score == nil {
		return APIErrorNotFound("Score")
	}

	if score.Failed {
		return APIErrorBadRequest("Failed scores do not have replay data.")
	}

	// Cache Map
	quaPath, err := files.CacheQuaFile(score.Map)

	if err != nil {
		return APIErrorServerError("Failed to cache qua file", err)
	}

	// Cache Replay
	replayPath, err := files.CacheReplay(score.Id)

	if err != nil {
		return APIErrorServerError("Failed to cache replay file", err)
	}

	return nil
}
