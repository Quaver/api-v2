package handlers

import (
	"net/http"
	"strings"

	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
)

// ClearScoreboardCache clears every cached scoreboard for a canonical map MD5.
func ClearScoreboardCache(c *gin.Context) *APIError {
	md5 := strings.ToLower(c.Param("md5"))

	deleted, err := db.ClearScoreboardCache(c.Request.Context(), md5)
	if err != nil {
		return APIErrorServerError("Error clearing scoreboard cache", err)
	}

	c.JSON(http.StatusOK, gin.H{"map_md5": md5, "deleted": deleted})
	return nil
}
