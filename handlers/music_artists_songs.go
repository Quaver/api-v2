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

// UpdateMusicArtistSong Updates metadata for a given song
// Endpoint: POST /v2/artists/song/:id
func UpdateMusicArtistSong(c *gin.Context) *APIError {
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
		Name   *string `form:"name" json:"name"`
		BPM    *int    `form:"bpm" json:"bpm"`
		Length *int    `form:"length" json:"length"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	song, err := db.GetMusicArtistSongById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving song from db", err)
	}

	if song == nil {
		return APIErrorNotFound("Song")
	}

	if body.Name != nil {
		if err := song.UpdateName(*body.Name); err != nil {
			return APIErrorServerError("Error updating song name", err)
		}
	}

	if body.BPM != nil {
		if err := song.UpdateBPM(*body.BPM); err != nil {
			return APIErrorServerError("Error updating song BPM", err)
		}
	}

	if body.Length != nil {
		if err := song.UpdateLength(*body.Length); err != nil {
			return APIErrorServerError("Error updating song length", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "The song has been successfully updated."})
	return nil
}

// DeleteMusicArtistSong Deletes a music artist's song
// Endpoint: DELETE /v2/artists/song/:id
func DeleteMusicArtistSong(c *gin.Context) *APIError {
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

	song, err := db.GetMusicArtistSongById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving song from db", err)
	}

	if song == nil {
		return APIErrorNotFound("Song")
	}

	_ = cache.RemoveCacheServerMusicArtistSong(song.Id)

	if err := azure.Client.DeleteBlob("music-artist-songs", fmt.Sprintf("%v.mp3", song.Id)); err != nil {
		return APIErrorServerError("Error deleting song from  azure", err)
	}

	if err := db.SQL.Delete(&db.MusicArtistSong{}, "id = ?", song.Id).Error; err != nil {
		return APIErrorServerError("Error deleting song from db", err)
	}

	if err := db.SyncMusicArtistSongSortOrders(song.AlbumId); err != nil {
		return APIErrorServerError("Error syncing song sort orders", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The song has been successfully deleted."})
	return nil
}

// SortMusicArtistSongs Sort's an album's songs
// Endpoint: POST /v2/artists/album/:id/song/sort
func SortMusicArtistSongs(c *gin.Context) *APIError {
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

	songs, err := db.GetMusicArtistSongsInAlbum(id)

	if err != nil {
		return APIErrorServerError("Error retrieving songs in album", err)
	}

	if len(songs) == 0 {
		return APIErrorBadRequest("There are no songs to sort.")
	}

	err = db.CustomizeSortOrder(songs, body.Ids, func(song *db.MusicArtistSong, sortOrder int) error {
		return song.UpdateSortOrder(sortOrder)
	}, func() error {
		return db.SyncMusicArtistSongSortOrders(songs[0].AlbumId)
	})

	if err != nil {
		return APIErrorServerError("Error sorting albums", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The albums have been successfully sorted."})
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
