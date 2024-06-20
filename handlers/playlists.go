package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// GetPlaylist Gets an individual playlist
// Endpoint: GET /v2/playlists/:id
func GetPlaylist(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	playlist, err := db.GetPlaylist(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving playlist from db", err)
	}

	if playlist == nil {
		return APIErrorNotFound("Playlist")
	}

	c.JSON(http.StatusOK, gin.H{"playlist": playlist})
	return nil
}
