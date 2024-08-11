package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

// InsertMusicArtist Creates a new music artists
// Endpoint: POST /artists
func InsertMusicArtist(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !canUserAccessAdminRoute(c) {
		return nil
	}

	body := struct {
		Name        string `form:"name" json:"name" binding:"required"`
		Description string `form:"description" json:"description"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	artist := db.MusicArtist{
		Name:        body.Name,
		Description: body.Description,
		SortOrder:   0,
		Visible:     true,
	}

	if err := artist.Insert(); err != nil {
		return APIErrorServerError("Error inserting music artist", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"artist":  artist,
		"message": "The music artist has been successfully created.",
	})
	return nil
}

// GetMusicArtists Returns all music artists
// Endpoint: GET /artists
func GetMusicArtists(c *gin.Context) *APIError {
	artists, err := db.GetMusicArtists()

	if err != nil {
		return APIErrorServerError("Error retrieving music artists", err)
	}

	c.JSON(http.StatusOK, gin.H{"music_artists": artists})
	return nil
}

// UpdateMusicArtist Updates the name, description, and links for a music artists
// Endpoint: POST /artists/:id
func UpdateMusicArtist(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !canUserAccessAdminRoute(c) {
		return nil
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	body := struct {
		Name          *string `form:"name" json:"name"`
		Description   *string `form:"description" json:"description"`
		ExternalLinks *string `form:"external_links" json:"external_links"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	artist, err := db.GetMusicArtistById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving artist by id", err)
	}

	if artist == nil {
		return APIErrorNotFound("Artist")
	}

	if body.Name != nil {
		if err := artist.UpdateName(*body.Name); err != nil {
			return APIErrorServerError("Error updating artist name", err)
		}
	}

	if body.Description != nil {
		if err := artist.UpdateDescription(*body.Description); err != nil {
			return APIErrorServerError("Error updating artist description", err)
		}
	}

	if body.ExternalLinks != nil {
		if err := artist.UpdateExternalLinks(*body.ExternalLinks); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "json") {
				return APIErrorBadRequest("You have provided invalid JSON for `external_links`")
			}

			return APIErrorServerError("Error updating artist external links", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "The music artist has been successfully updated."})
	return nil
}
