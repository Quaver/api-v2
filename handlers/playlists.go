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

	playlist, err := db.GetPlaylistFull(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving playlist from db", err)
	}

	if playlist == nil {
		return APIErrorNotFound("Playlist")
	}

	c.JSON(http.StatusOK, gin.H{"playlist": playlist})
	return nil
}

// UpdatePlaylist Updates a playlists name/description
// Endpoint: POST /v2/playlists/:id/update
func UpdatePlaylist(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Name        string `form:"name" json:"name"`
		Description string `form:"description" json:"description"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	playlist, err := db.GetPlaylist(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving playlist from db", err)
	}

	if playlist == nil {
		return APIErrorNotFound("Playlist")
	}

	if playlist.UserId != user.Id {
		return APIErrorForbidden("You do not own this playlist.")
	}

	if len(body.Name) > 0 {
		playlist.Name = body.Name
	}

	if len(body.Description) > 0 {
		playlist.Description = body.Description
	}

	if err := db.SQL.Save(&playlist).Error; err != nil {
		return APIErrorServerError("Error updating playlist in db", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your playlist has been successfully updated"})
	return nil
}
