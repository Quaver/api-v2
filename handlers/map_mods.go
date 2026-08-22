package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// GetMapMods Gets mods for a given map
// Endpoint: GET /v2/maps/:id/mods
func GetMapMods(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	mods, err := db.GetMapMods(id)

	if err != nil {
		return APIErrorServerError("Error retrieving map mods from db", err)
	}

	c.JSON(http.StatusOK, gin.H{"mods": mods})
	return nil
}

// SubmitMapMod Inserts a map mod to the db
// Endpoint: POST /v2/maps/:id/mods
func SubmitMapMod(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Type         db.MapModType `form:"type" json:"type" binding:"required"`
		MapTimestamp *string       `form:"map_timestamp" json:"map_timestamp"`
		Comment      string        `form:"comment" json:"comment" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if body.Type != db.ModTypeIssue && body.Type != db.ModTypeSuggestion {
		return APIErrorBadRequest("You have provided an invalid mod type")
	}

	if body.MapTimestamp != nil {
		if len(*body.MapTimestamp) > 5000 {
			return APIErrorBadRequest("The map timestamp can't be greater than 5,000 characters")
		}

		if !isMapTimestampValid(*body.MapTimestamp) {
			return APIErrorBadRequest("You have provided an invalid map timestamp")
		}
	}

	if len(body.Comment) > 5000 {
		return APIErrorBadRequest("Your comment can't be greater than 5,000 characters")
	}

	songMap, err := db.GetMapById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving map from db", err)
	}

	if songMap == nil {
		return APIErrorNotFound("Map")
	}

	if songMap.RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("You cannot submit mods for a map that is already ranked.")
	}

	mod := &db.MapMod{
		Type:         body.Type,
		MapId:        songMap.Id,
		AuthorId:     user.Id,
		MapTimestamp: body.MapTimestamp,
		Comment:      body.Comment,
	}

	if err := mod.Insert(); err != nil {
		return APIErrorServerError("Error inserting mod to db", err)
	}

	if err := db.NewMapModNotification(songMap, mod).Insert(); err != nil {
		return APIErrorServerError("Error inserting map mod notification", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your mod has been successfully added."})
	return nil
}

// UpdateMapModStatus Updates the status of a map mod
// Endpoint: POST /v2/maps/:id/mods/:mod_id/status
func UpdateMapModStatus(c *gin.Context) *APIError {
	modId, err := strconv.Atoi(c.Param("mod_id"))

	if err != nil {
		return APIErrorBadRequest("Invalid mod id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Status db.MapModStatus `form:"status" json:"status" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if body.Status != db.ModStatusPending &&
		body.Status != db.ModStatusAccepted &&
		body.Status != db.ModStatusDenied {
		return APIErrorBadRequest("You have provided an invalid mod status")
	}

	mod, err := db.GetModById(modId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mod from database", err)
	}

	if mod == nil {
		return APIErrorNotFound("Mod")
	}

	songMap, err := db.GetMapById(mod.MapId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving map from database", err)
	}

	if mod == nil {
		return APIErrorNotFound("Map")
	}

	if songMap.CreatorId != user.Id {
		return APIErrorForbidden("This map does not belong to you.")
	}

	if err := mod.UpdateStatus(body.Status); err != nil {
		return APIErrorServerError("Error updating mod status in database", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "The mod status has been successfully updated."})
	return nil
}

// SubmitMapModComment Submits a comment for a map mod
// Endpoint: POST /v2/maps/:id/mods/:mod_id/comment
func SubmitMapModComment(c *gin.Context) *APIError {
	modId, err := strconv.Atoi(c.Param("mod_id"))

	if err != nil {
		return APIErrorBadRequest("Invalid mod id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Comment string `form:"comment" json:"comment" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if len(body.Comment) > 5000 {
		return APIErrorBadRequest("Your comment must not be greater than 5,000 characters.")
	}

	mod, err := db.GetModById(modId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mod from database", err)
	}

	if mod == nil {
		return APIErrorNotFound("Mod")
	}

	mapQua, err := db.GetMapById(mod.MapId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error getting map by id", err)
	}

	if mapQua == nil {
		return APIErrorNotFound("Map")
	}

	comment := &db.MapModComment{
		MapModId: modId,
		AuthorId: user.Id,
		Comment:  body.Comment,
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting map mod comment into database", err)
	}

	if err := db.NewMapModCommentNotification(mapQua, mod, comment).Insert(); err != nil {
		return APIErrorServerError("Error inserting map mod comment notification", err)
	}

	if mod.AuthorId != mapQua.CreatorId && comment.AuthorId != mapQua.CreatorId {
		notif := db.NewMapModCommentNotification(mapQua, mod, comment)
		notif.ReceiverId = mapQua.CreatorId

		if err := notif.Insert(); err != nil {
			return APIErrorServerError("Error inserting map mod comment notif for map creator", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your comment has been successfully added."})
	return nil
}

// EditMapMod Edits the content of a pending map mod.
// Endpoint: POST /v2/map/mod/:id/edit
func EditMapMod(c *gin.Context) *APIError {
	modId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid mod id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		MapTimestamp *string `form:"map_timestamp" json:"map_timestamp"`
		Comment      string  `form:"comment" json:"comment" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if apiErr := validateEditableMapModComment(body.Comment); apiErr != nil {
		return apiErr
	}

	var apiErr *APIError
	body.MapTimestamp, apiErr = normalizeEditableMapModTimestamp(body.MapTimestamp)

	if apiErr != nil {
		return apiErr
	}

	mod, err := db.GetModById(modId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving map mod from database", err)
	}

	if mod == nil {
		return APIErrorNotFound("Mod")
	}

	if mod.AuthorId != user.Id {
		return APIErrorForbidden("You are not the author of this mod.")
	}

	if mod.Status != db.ModStatusPending {
		return APIErrorForbidden("You can only edit pending mods.")
	}

	if err := mod.Edit(body.Comment, body.MapTimestamp); err != nil {
		return APIErrorServerError("Error updating map mod in the database", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your mod has been successfully edited."})
	return nil
}

// EditMapModComment Edits the content of a comment on a pending map mod.
// Endpoint: POST /v2/map/mod/comment/:id/edit
func EditMapModComment(c *gin.Context) *APIError {
	commentId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid comment id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Comment string `form:"comment" json:"comment" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if apiErr := validateEditableMapModComment(body.Comment); apiErr != nil {
		return apiErr
	}

	comment, err := db.GetMapModCommentById(commentId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving map mod comment from database", err)
	}

	if comment == nil {
		return APIErrorNotFound("Comment")
	}

	if comment.AuthorId != user.Id {
		return APIErrorForbidden("You are not the author of this comment.")
	}

	mod, err := db.GetModById(comment.MapModId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving map mod from database", err)
	}

	if mod == nil {
		return APIErrorNotFound("Mod")
	}

	if mod.Status != db.ModStatusPending {
		return APIErrorForbidden("You can only edit comments on pending mods.")
	}

	if err := comment.Edit(body.Comment); err != nil {
		return APIErrorServerError("Error updating map mod comment in the database", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your comment has been successfully edited."})
	return nil
}

func validateEditableMapModComment(comment string) *APIError {
	if len(comment) == 0 || len(comment) > 5000 {
		return APIErrorBadRequest("Your comment must be between 1 and 5,000 characters")
	}

	return nil
}

func normalizeEditableMapModTimestamp(timestamp *string) (*string, *APIError) {
	if timestamp == nil || len(*timestamp) == 0 {
		return nil, nil
	}

	if len(*timestamp) > 5000 {
		return nil, APIErrorBadRequest("The map timestamp can't be greater than 5,000 characters")
	}

	if !isMapTimestampValid(*timestamp) {
		return nil, APIErrorBadRequest("You have provided an invalid map timestamp")
	}

	return timestamp, nil
}

// Returns if a map timestamp has valid syntax
// Time OR Time|Lane,Time|Lane,...
func isMapTimestampValid(str string) bool {
	if len(str) == 0 {
		return true
	}

	// Check if string has just the time
	if match, _ := regexp.MatchString(`^\d+$`, str); match == true {
		return true
	}

	// Check if string is in Time|Lane format
	for _, timestamp := range strings.Split(str, ",") {
		if match, _ := regexp.MatchString(`\d+\|\d+`, timestamp); match == false {
			return false
		}
	}

	return true
}
