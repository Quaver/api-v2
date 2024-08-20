package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateMusicArtistAlbum Creates a new album for a music artist
// Endpoint: POST /v2/artists/:id/album
func CreateMusicArtistAlbum(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !canUserAccessAdminRoute(c) {
		return nil
	}

	body := struct {
		Name string `form:"name" json:"name" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	artist, apiErr := getMusicArtistFromParams(c)

	if apiErr != nil {
		return apiErr
	}

	albums, err := db.GetMusicArtistAlbums(artist.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving albums from db", err)
	}

	album := &db.MusicArtistAlbum{
		ArtistId:  artist.Id,
		Name:      body.Name,
		SortOrder: len(albums),
	}

	if err := db.SQL.Create(&album).Error; err != nil {
		return APIErrorServerError("Error inserting album in db", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"album":   album,
		"message": "The music artist has been successfully created.",
	})

	return nil
}
