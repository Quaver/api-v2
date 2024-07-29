package handlers

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/tools"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// GetVirtualReplayPlayerOutput Plays the virtual replay player & returns the output
// Endpoint: GET /v2/scores/:id/stats
func GetVirtualReplayPlayerOutput(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("md5")) // Have to name the id as MD5 due to gin limitation...

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

	quaPath, err := files.CacheQuaFile(score.Map)

	if err != nil {
		return APIErrorServerError("Failed to cache qua file", err)
	}

	replayPath, err := files.CacheReplay(score.Id)

	if err != nil {
		return APIErrorServerError("Failed to cache replay file", err)
	}

	var data interface{}
	key := fmt.Sprintf("quaver:score:%v:stats", score.Id)

	err = db.CacheJsonInRedis(key, &data, time.Hour*1, false, func() error {
		replayStats, err := tools.PlayReplayVirtually(quaPath, replayPath, score.Modifiers)

		if err != nil {
			return err
		}

		data = map[string]interface{}{
			"stats": map[string]interface{}{
				"score":           replayStats.Score,
				"accuracy":        replayStats.Accuracy,
				"max_combo":       replayStats.MaxCombo,
				"count_marvelous": replayStats.CountMarv,
				"count_perfect":   replayStats.CountPerf,
				"count_great":     replayStats.CountGreat,
				"count_good":      replayStats.CountGood,
				"count_okay":      replayStats.CountOkay,
				"count_miss":      replayStats.CountMiss,
			},
			"deviance": replayStats.Hits,
		}

		return nil
	})

	if err != nil {
		return APIErrorServerError("Error getting replay stats", err)
	}

	c.JSON(http.StatusOK, data)
	return nil
}
