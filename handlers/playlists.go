package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// CreatePlaylist Creates a new playlist
// Endpoint: /v2/playlists
func CreatePlaylist(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Name        string `form:"name" json:"name" binding:"required"`
		Description string `form:"description" json:"description" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if len(body.Name) > 100 {
		return APIErrorBadRequest("Your playlist name cannot be longer than 100 characters.")
	}

	if len(body.Description) > 2000 {
		return APIErrorBadRequest("Your playlist description cannot be longer than 2000 characters.")
	}

	playlist := db.Playlist{
		UserId:      user.Id,
		Name:        body.Name,
		Description: body.Description,
	}

	if err := playlist.Insert(); err != nil {
		return APIErrorServerError("Error inserting playlist into db", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "You have successfully created a new playlist",
		"playlist": playlist,
	})

	return nil
}

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

// DeletePlaylist Deletes (hides) a playlist
// Endpoint: DELETE /v2/playlists/:id
func DeletePlaylist(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
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

	playlist.Visible = false

	if err := db.SQL.Save(&playlist).Error; err != nil {
		return APIErrorServerError("Error deleting (updating visibility) playlist in db", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your playlist has been successfully deleted"})
	return nil
}

// GetUserPlaylists Returns a user's created playlists
// Endpoint: /v2/user/:id/playlists
func GetUserPlaylists(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	playlists, err := db.GetUserPlaylists(id)

	if err != nil {
		return APIErrorServerError("Error retrieving user playlists", err)
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
	return nil
}