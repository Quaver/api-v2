package handlers

import (
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/cache"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MusicArtistImage int

const (
	MusicArtistImageAvatar MusicArtistImage = iota
	MusicArtistImageBanner
)

// UploadMusicArtistAvatar Uploads a music artist's avatar
// Endpoint: POST /v2/artists/:id/avatar
func UploadMusicArtistAvatar(c *gin.Context) *APIError {
	return uploadMusicArtistImage(c, MusicArtistImageAvatar)
}

// UploadMusicArtistBanner Uploads a music artist's banner
// Endpoint: POST /v2/artists/:id/banner
func UploadMusicArtistBanner(c *gin.Context) *APIError {
	return uploadMusicArtistImage(c, MusicArtistImageBanner)
}

// Uploads a music artist's image (avatar/banner)
func uploadMusicArtistImage(c *gin.Context, image MusicArtistImage) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !canUserAccessAdminRoute(c) {
		return nil
	}

	artist, apiErr := getMusicArtistFromParams(c)

	if apiErr != nil {
		return apiErr
	}

	file, apiErr := validateUploadedImage(c)

	if apiErr != nil {
		return apiErr
	}

	var err error
	fileName := fmt.Sprintf("%v.jpg", artist.Id)

	switch image {
	case MusicArtistImageAvatar:
		_ = cache.RemoveCacheServerMusicArtistAvatar(artist.Id)
		err = azure.Client.UploadFile("music-artist-avatars", fileName, file)
		break
	case MusicArtistImageBanner:
		_ = cache.RemoveCacheServerMusicArtistBanner(artist.Id)
		err = azure.Client.UploadFile("music-artist-banners", fileName, file)
		break
	}

	if err != nil {
		return APIErrorServerError("Error uploading music artist file", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your music artist image has been successfully uploaded."})
	return nil
}
