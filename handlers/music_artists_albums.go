package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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

// UpdateMusicArtistAlbum Updates  a music artist album
// Endpoint: POST /v2/artists/album/:id
func UpdateMusicArtistAlbum(c *gin.Context) *APIError {
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
		Name string `form:"name" json:"name" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	album, err := db.GetMusicArtistAlbumById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving album from db", err)
	}

	if album == nil {
		return APIErrorNotFound("Album")
	}

	if err := album.UpdateName(body.Name); err != nil {
		return APIErrorServerError("Error updating album name", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The music artist has been updated."})
	return nil
}

// DeleteMusicArtistAlbum Deletes a music artist's album
func DeleteMusicArtistAlbum(c *gin.Context) *APIError {
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

	album, err := db.GetMusicArtistAlbumById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving album from db", err)
	}

	if album == nil {
		return APIErrorNotFound("Album")
	}

	if err := album.Delete(); err != nil {
		return APIErrorServerError("Error deleting album", err)
	}

	if err := db.SyncMusicArtistAlbumSortOrders(album.ArtistId); err != nil {
		return APIErrorServerError("Error syncing album sort orders", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The music artist has been successfully deleted."})
	return nil
}

// SortMusicArtistAlbums Sorts music artist albums
// Endpoints: POST: /v2/artists/:id/album/sort
func SortMusicArtistAlbums(c *gin.Context) *APIError {
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
		Ids []int `form:"ids" json:"ids" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	albums, err := db.GetMusicArtistAlbums(id)

	if err != nil {
		return APIErrorServerError("Error retrieving albums", err)
	}

	err = db.CustomizeSortOrder(albums, body.Ids, func(album *db.MusicArtistAlbum, sortOrder int) error {
		return album.UpdateSortOrder(sortOrder)
	}, func() error {
		return db.SyncMusicArtistSortOrders()
	})

	if err != nil {
		return APIErrorServerError("Error sorting albums", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The albums have been successfully sorted."})
	return nil
}
