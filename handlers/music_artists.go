package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
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
