package handlers

import (
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/cache"
	"github.com/Quaver/api2/db"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// UploadMusicArtistSong Uploads a new song for a music artist's album
// Endpoint: POST /v2/artists/album/:id/song
func UploadMusicArtistSong(c *gin.Context) *APIError {
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
		Name   string `form:"name" json:"name" binding:"required"`
		BPM    int    `form:"bpm" json:"bpm" binding:"required"`
		Length int    `form:"length" json:"length" binding:"required"`
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

	fileBytes, apiErr := validateUploadedAudio(c)

	if apiErr != nil {
		return apiErr
	}

	song := db.MusicArtistSong{
		AlbumId:   id,
		Name:      body.Name,
		Length:    body.Length,
		BPM:       body.BPM,
		SortOrder: len(album.Songs),
	}

	if err := db.SQL.Create(&song).Error; err != nil {
		return APIErrorServerError("Error inserting song into db", err)
	}

	_ = cache.RemoveCacheServerMusicArtistSong(song.Id)

	if err := azure.Client.UploadFile("music-artist-songs", fmt.Sprintf("%v.mp3", song.Id), fileBytes); err != nil {
		return APIErrorServerError("Error uploading song to azure", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Your song has been successfully uploaded.",
		"song":    song,
	})

	return nil
}

// Validates an uploaded audio file and returns the file bytes
func validateUploadedAudio(c *gin.Context) ([]byte, *APIError) {
	fileHeader, _ := c.FormFile("audio")

	if fileHeader == nil {
		return nil, APIErrorBadRequest("You must provide a valid `audio` file.")
	}

	file, err := fileHeader.Open()

	if err != nil {
		return nil, APIErrorServerError("Error opening file", err)
	}

	defer file.Close()

	var fileBytes = make([]byte, fileHeader.Size)

	if _, err = file.Read(fileBytes); err != nil {
		return nil, APIErrorServerError("Error reading file", err)
	}

	if mime := mimetype.Detect(fileBytes); mime == nil || mime.String() != "audio/mpeg" {
		return nil, APIErrorBadRequest("The file you provided was not a valid .mp3 file.")
	}

	return fileBytes, nil
}
