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

	comment := &db.MapModComment{
		MapModId: modId,
		AuthorId: user.Id,
		Comment:  body.Comment,
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting map mod comment into database", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your comment has been successfully added."})
	return nil
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
