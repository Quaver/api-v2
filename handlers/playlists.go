package handlers

import (
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/cache"
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
		if err := playlist.UpdateName(body.Name); err != nil {
			return APIErrorServerError("Error updating playlist name", err)
		}
	}

	if len(body.Description) > 0 {
		if err := playlist.UpdateDescription(body.Description); err != nil {
			return APIErrorServerError("Error updating playlist description", err)
		}
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

	if err := playlist.UpdateVisibility(false); err != nil {
		return APIErrorServerError("Error updating playlist visibility", err)
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

	if _, apiErr := getUserById(id, canAuthedUserViewBannedUsers(c)); apiErr != nil {
		return apiErr
	}

	playlists, err := db.GetUserPlaylists(id)

	if err != nil {
		return APIErrorServerError("Error retrieving user playlists", err)
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
	return nil
}

// GetPlaylistContainsMap Returns if a playlist contains an individual map id
// Endpoint: /v2/playlists/:id/contains/:map_id
func GetPlaylistContainsMap(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mapId, err := strconv.Atoi(c.Param("map_id"))

	if err != nil {
		return APIErrorBadRequest("Invalid map_id")
	}

	exists, err := db.DoesPlaylistContainMap(id, mapId)

	if err != nil {
		return APIErrorServerError("Error checking if map exists in playlist", err)
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
	return nil
}

// SearchPlaylists Searches for playlists by name/creator username
// Endpoint: /v2/playlists/search?query=
func SearchPlaylists(c *gin.Context) *APIError {
	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	query := c.Query("query")

	playlists, err := db.SearchPlaylists(query, 50, page)

	if err != nil {
		return APIErrorServerError("Error searching playlists", err)
	}

	totalCount, err := db.GetTotalPlaylistCount(query)

	if err != nil {
		return APIErrorServerError("Error retrieving total playlist count", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_playlist_count": totalCount,
		"playlists":            playlists,
	})

	return nil
}

// Struct reused when parsing/fetching data to use in endpoints where we're adding/removing maps to playlists
type addRemoveMapPlaylistData struct {
	User     *db.User
	Playlist *db.Playlist
	Map      *db.MapQua
}

// Parses any ids, performs validation and returns data to be used when adding/removing maps from playlists
func validateAddRemoveMapFromPlaylist(c *gin.Context) (*addRemoveMapPlaylistData, *APIError) {
	playlistId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return nil, APIErrorBadRequest("Invalid id")
	}

	mapId, err := strconv.Atoi(c.Param("map_id"))

	if err != nil {
		return nil, APIErrorBadRequest("Invalid map_id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil, APIErrorUnauthorized("User not authenticated")
	}

	playlist, err := db.GetPlaylistFull(playlistId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, APIErrorServerError("Error retrieving playlist from database", err)
	}

	if playlist == nil {
		return nil, APIErrorNotFound("Playlist")
	}

	if playlist.UserId != user.Id {
		return nil, APIErrorForbidden("You do not own this playlist.")
	}

	songMap, err := db.GetMapById(mapId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, APIErrorServerError("Error getting map from database", err)
	}

	if songMap == nil {
		return nil, APIErrorNotFound("Map")
	}

	if songMap.MapsetId == -1 {
		return nil, APIErrorBadRequest("You cannot add/remove this map to your playlist.")
	}

	return &addRemoveMapPlaylistData{
		User:     user,
		Playlist: playlist,
		Map:      songMap,
	}, nil
}

// AddMapToPlaylist Adds a map to a playlist
// Endpoint: /v2/playlist/:id/add/:map_id
func AddMapToPlaylist(c *gin.Context) *APIError {
	data, apiErr := validateAddRemoveMapFromPlaylist(c)

	if apiErr != nil {
		return apiErr
	}

	var existingMapset *db.PlaylistMapset

	for _, mapset := range data.Playlist.Mapsets {
		for _, playlistMap := range mapset.Maps {
			if playlistMap.MapId == data.Map.Id {
				return APIErrorBadRequest("This map already exists in the playlist.")
			}
		}

		// Set existing mapset
		if mapset.MapsetId == data.Map.MapsetId {
			existingMapset = mapset
		}
	}

	// Create new playlist mapset
	if existingMapset == nil {
		existingMapset = &db.PlaylistMapset{
			PlaylistId: data.Playlist.Id,
			MapsetId:   data.Map.MapsetId,
		}

		if err := existingMapset.Insert(); err != nil {
			return APIErrorServerError("Error inserting playlist mapset to database", err)
		}
	}

	playlistMap := &db.PlaylistMap{
		PlaylistId:        data.Playlist.Id,
		MapId:             data.Map.Id,
		PlaylistsMapsetId: existingMapset.Id,
	}

	if err := playlistMap.Insert(); err != nil {
		return APIErrorServerError("Error inserting playlist map in database", err)
	}

	if err := db.UpdatePlaylistMapCount(data.Playlist.Id, data.Playlist.MapCount+1); err != nil {
		return APIErrorServerError("Error updating playlist map count (add)", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The map has been successfully added to your playlist."})
	return nil
}

// RemoveMapFromPlaylist Removes a map from a playlist
// Endpoint: /v2/playlist/:id/remove/:map_id
func RemoveMapFromPlaylist(c *gin.Context) *APIError {
	data, apiErr := validateAddRemoveMapFromPlaylist(c)

	if apiErr != nil {
		return apiErr
	}

	var deleteMap bool
	var deleteMapset bool

	for _, playlistMapset := range data.Playlist.Mapsets {
		for _, playlistMap := range playlistMapset.Maps {
			if playlistMap.MapId == data.Map.Id {
				deleteMap = true
			}

			if len(playlistMapset.Maps) == 1 {
				deleteMapset = true
			}
		}
	}

	if !deleteMap {
		return APIErrorBadRequest("This map is not in your playlist.")
	}

	if err := db.DeletePlaylistMap(data.Playlist.Id, data.Map.Id); err != nil {
		return APIErrorServerError("Error removing playlist map from db", err)
	}

	if deleteMapset {
		if err := db.DeletePlaylistMapset(data.Playlist.Id, data.Map.MapsetId); err != nil {
			return APIErrorServerError("Error removing playlist mapset from db", err)
		}
	}

	if err := db.UpdatePlaylistMapCount(data.Playlist.Id, data.Playlist.MapCount-1); err != nil {
		return APIErrorServerError("Error updating playlist map count (remove)", err)
	}

	c.JSON(http.StatusOK, gin.H{"Message": "You have successfully removed the map from your playlist."})
	return nil
}

// UploadPlaylistCover Uploads a playlist cover
// Endpoint: POST /v2/playlists/:id/cover
func UploadPlaylistCover(c *gin.Context) *APIError {
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

	file, apiErr := validateUploadedImage(c)

	if apiErr != nil {
		return apiErr
	}

	_ = cache.RemoveCacheServerPlaylistCover(playlist.Id)

	if err := azure.Client.UploadFile("playlists", fmt.Sprintf("%v.jpg", playlist.Id), file); err != nil {
		return APIErrorServerError("Failed to upload file", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Your playlist cover has been successfully updated!",
	})

	return nil
}
