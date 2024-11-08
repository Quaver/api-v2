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

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	status, err := strconv.Atoi(c.Query("status"))

	if err != nil {
		status = enums.RankedStatusRanked
	}

	if _, apiErr := getUserById(id, canAuthedUserViewBannedUsers(c)); apiErr != nil {
		return apiErr
	}

	mapsets, err := db.GetUserMapsetsFiltered(id, enums.RankedStatus(status), page, 50)

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

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Failed to retrieve mapset from database", err)
	}

	if mapset == nil {
		return APIErrorNotFound("Mapset")
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

	if err := db.UpdateElasticSearchMapset(*mapset); err != nil {
		return APIErrorServerError("Failed to index ranked mapset in elastic search", err)
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
		playlist, err := db.GetPlaylist(id)

		if err != nil && err != gorm.ErrRecordNotFound {
			return APIErrorServerError("Error retrieving playlist from db", err)
		}

		if playlist == nil {
			continue
		}

		if err := db.DeletePlaylistMapset(playlistMapset.PlaylistId, playlistMapset.MapsetId); err != nil {
			return APIErrorServerError("Error deleting playlist mapset", err)
		}

		for _, playlistMap := range playlistMapset.Maps {
			if err := db.DeletePlaylistMap(playlistMapset.PlaylistId, playlistMap.MapId); err != nil {
				return APIErrorServerError("Error deleting playlist map", err)
			}
		}

		if err := db.UpdatePlaylistMapCount(playlistMapset.PlaylistId, playlist.MapCount-len(playlistMapset.Maps)); err != nil {
			return APIErrorServerError("Error updating playlist map count", err)
		}
	}

	err = db.DeleteMapset(mapset.Id)

	if err != nil {
		return APIErrorServerError("Error deleting mapset", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully deleted your mapset!"})
	return nil
}

// MarkMapsetAsExplicit Marks a mapset as explicit
// POST /v2/mapset/:id/explicit
func MarkMapsetAsExplicit(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasPrivilege(user.Privileges, enums.PrivilegeRankMapsets) {
		return APIErrorForbidden("You do not have permission to perform this action.")
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset data", err)
	}

	if mapset == nil {
		return APIErrorNotFound("Mapset")
	}

	if err := mapset.UpdateExplicit(true); err != nil {
		return APIErrorServerError("Error setting mapset as explicit", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully marked that mapset as explicit."})
	return nil
}

// MarkMapsetAsNotExplicit Marks a mapsets as not explicit
// POST /v2/mapset/:id/unexplicit
func MarkMapsetAsNotExplicit(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasPrivilege(user.Privileges, enums.PrivilegeRankMapsets) {
		return APIErrorForbidden("You do not have permission to perform this action.")
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset data", err)
	}

	if mapset == nil {
		return APIErrorNotFound("Mapset")
	}

	if err := mapset.UpdateExplicit(false); err != nil {
		return APIErrorServerError("Error setting mapset as explicit", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully marked that mapset as clean."})
	return nil
}

// UpdateElasticSearchMapset Updates a mapset in elastic search
// Endpoint: GET /v2/mapset/:id/elastic
func UpdateElasticSearchMapset(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !canUserAccessAdminRoute(c) {
		return APIErrorForbidden("You do not have permission to access this endpoint.")
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset in database", err)
	}

	if mapset == nil || !mapset.IsVisible {
		c.JSON(http.StatusOK, gin.H{"message": "Mapset not found, so skipping."})
		return nil
	}

	if err := db.UpdateElasticSearchMapset(*mapset); err != nil {
		return APIErrorServerError("Error updating mapset in elastic search", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The mapset has been successfully updated in ElasticSearch."})
	return nil
}

// GetMapsetsSearch Gets a list of mapsets that match a search query
// Endpoint: GET /v2/mapsets/search
func GetMapsetsSearch(c *gin.Context) *APIError {
	body := db.NewElasticMapsetSearchOptions()

	if err := c.BindQuery(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	body.BindAndValidate()

	mapsets, total, err := db.SearchElasticMapsets(body)

	if err != nil {
		return APIErrorServerError("Error retrieving mapsets from elastic search", err)
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "mapsets": mapsets})
	return nil
}
