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

	playlists, err := db.SearchPlaylists(c.Query("query"), 50, page)

	if err != nil {
		return APIErrorServerError("Error searching playlists", err)
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
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

	for _, playlistMapset := range data.Playlist.Mapsets {
		for _, playlistMap := range playlistMapset.Maps {
			// Map exists, so remove it
			if playlistMap.MapId == data.Map.Id {
				if err := db.SQL.Delete(&playlistMap).Error; err != nil {
					return APIErrorServerError("Error removing playlist map from db", err)
				}

				// This was the last map in the mapset, so remove the playlist mapset from the db
				if len(playlistMapset.Maps) == 1 {
					if err := db.SQL.Delete(&playlistMapset).Error; err != nil {
						return APIErrorServerError("Error removing playlist mapset from db", err)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "The map has been successfully removed from your playlist."})
	return nil
}
