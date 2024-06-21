package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// GetMapsetById Retrieves a mapset from the database by its id
// Endpoint: GET /v2/mapset/:id
func GetMapsetById(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mapset, err := db.GetMapsetById(id)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("Mapset")
	default:
		return APIErrorServerError("Failed to get mapset from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"mapset": mapset})
	return nil
}

// GetUserMapsets Gets a user's uploaded mapsets
// Endpoint: GET /v2/user/:id/mapsets
func GetUserMapsets(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mapsets, err := db.GetUserMapsets(id)

	if err != nil {
		return APIErrorServerError("Failed to get mapsets from database", err)
	}

	c.JSON(http.StatusOK, gin.H{"mapsets": mapsets})
	return nil
}

// UpdateMapsetDescription Updates the description of a given mapset
// Endpoint: PATCH /v2/mapset/:id/description
func UpdateMapsetDescription(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil {
		return APIErrorServerError("Failed to retrieve mapset from database", err)
	}

	if mapset.CreatorID != user.Id {
		return APIErrorForbidden("You are not the owner of this mapset.")
	}

	body := struct {
		Description string `form:"description" json:"description"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if len(body.Description) > 2000 {
		return APIErrorBadRequest("Your mapset description cannot exceed 2,000 characters.")
	}

	err = db.UpdateMapsetDescription(mapset.Id, body.Description)

	if err != nil {
		return APIErrorServerError("Error updating mapset description", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your mapset description was successfully updated!"})
	return nil
}

// GetRankedMapsetIds Retrieves the list of ranked mapset ids
// Endpoint: GET /v2/mapset/ranked
func GetRankedMapsetIds(c *gin.Context) *APIError {
	mapsets, err := db.GetRankedMapsetIds()

	if err != nil {
		return APIErrorServerError("Error retrieving ranked mapset ids", err)
	}

	c.JSON(http.StatusOK, gin.H{"ranked_mapsets": mapsets})
	return nil
}

// GetMapsetOnlineOffsets Retrieves online offsets for all ranked mapsets
// Endpoint: GET /v2/mapset/offsets
func GetMapsetOnlineOffsets(c *gin.Context) *APIError {
	offsets, err := db.GetMapsetOnlineOffsets()

	if err != nil {
		return APIErrorServerError("Error retrieving mapset online offsets", err)
	}

	c.JSON(http.StatusOK, gin.H{"online_offsets": offsets})
	return nil
}

// DeleteMapset Removes a mapset from being visible on the server
// Endpoint: POST /v2/mapset/:id/delete
func DeleteMapset(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset data", err)
	}

	if mapset == nil {
		return APIErrorNotFound("Mapset")
	}

	if mapset.CreatorID != user.Id {
		return APIErrorForbidden("You are not the owner of this mapset.")
	}

	if !mapset.IsVisible {
		return APIErrorForbidden("This mapset has already been deleted.")
	}

	if mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("You cannot delete a ranked mapset.")
	}

	rankingQueue, err := db.GetRankingQueueMapset(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset from ranking queue", err)
	}

	if rankingQueue != nil && (rankingQueue.Status != db.RankingQueueDenied && rankingQueue.Status != db.RankingQueueBlacklisted) {
		return APIErrorForbidden("You cannot delete a mapset that is pending in the ranking queue.")
	}

	// Find all playlists that have this mapset in it and delete it from them
	playlistMapsets, err := db.GetPlaylistMapsetsByMapsetId(id)

	if err != nil {
		return APIErrorServerError("Error retrieving playlist mapsets by mapset id", err)
	}

	for _, playlistMapset := range playlistMapsets {
		if err := db.DeletePlaylistMapset(playlistMapset.PlaylistId, id); err != nil {
			return APIErrorServerError("Error deleting playlist mapset", err)
		}

		for _, playlistMap := range playlistMapset.Maps {
			if err := db.DeletePlaylistMap(playlistMapset.PlaylistId, playlistMap.MapId); err != nil {
				return APIErrorServerError("Error deleting playlist mapset", err)
			}
		}
	}

	err = db.DeleteMapset(mapset.Id)

	if err != nil {
		return APIErrorServerError("Error deleting mapset", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully deleted your mapset!"})
	return nil
}
