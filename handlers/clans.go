package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/stringutil"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

const (
	errClanUserCantJoin        string = "You must not already be in a clan to perform this action."
	errClanNameInvalid         string = "Your clan `name` must be between 3 and 30 characters and must contain only letters or numbers."
	errClanTagInvalid          string = "Your clan `tag` must be between 1 and 4 characters, must only contain letters or numbers, and must not be inappropriate."
	errClanAboutMeInvalid      string = "Your clan `about_me` must be between 0 and 2000 characters."
	errClanFavoriteModeInvalid string = "Your clan `favorite_mode` must be a valid mode id."
	errClanAccentColorInvalid  string = "Your clan `accent_color` must be a valid hex code."
	errClanNameExists          string = "A clan with that name already exists. Please choose a different name."
	errClanTagExists           string = "A clan with that tag already exists. Please choose a different tag."
)

// CreateClan Creates a new clan if the user is eligible to.
// Endpoint: POST /v2/clan
func CreateClan(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !user.CanJoinClan() {
		return APIErrorBadRequest(errClanUserCantJoin)
	}

	body := struct {
		Name string `form:"name" json:"name" binding:"required"`
		Tag  string `form:"tag" json:"tag" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if !db.IsValidClanName(body.Name) {
		return APIErrorBadRequest(errClanNameInvalid)
	}

	if !db.IsValidClanTag(body.Tag) {
		return APIErrorBadRequest(errClanTagInvalid)
	}

	nameExists, err := db.DoesClanExistByName(body.Name)

	if err != nil {
		return APIErrorServerError("Error checking if clan exists by name", err)
	}

	if nameExists {
		return APIErrorBadRequest(errClanNameExists)
	}

	tagExists, _, err := db.DoesClanExistByTag(body.Tag)

	if err != nil {
		return APIErrorServerError("Error checking if clan exists by tag", err)
	}

	if tagExists {
		return APIErrorBadRequest(errClanTagExists)
	}

	clan := db.Clan{
		OwnerId: user.Id,
		Name:    body.Name,
		Tag:     body.Tag,
	}

	if err := clan.Insert(); err != nil {
		return APIErrorServerError("Error inserting clan into database", err)
	}

	if err := db.UpdateUserClan(user.Id, clan.Id); err != nil {
		return APIErrorServerError("Error updating clan", err)
	}

	if err := db.DeleteUserClanInvites(user.Id); err != nil {
		return APIErrorServerError("Error deleting user clan invites", err)
	}

	if err := db.NewClanActivity(clan.Id, db.ClanActivityCreated, user.Id).Insert(); err != nil {
		return APIErrorServerError("Error inserting clan activity", err)
	}

	if err := db.UpdateAllClanLeaderboards(&clan); err != nil {
		return APIErrorServerError("Error adding clan to leaderboard", err)
	}

	c.JSON(http.StatusOK, struct {
		Message string   `json:"message"`
		Clan    *db.Clan `json:"clan"`
	}{
		Message: "Your clan has been successfully created.",
		Clan:    &clan,
	})

	logrus.Debugf("%v (#%v) has created the clan `%v` (#%v).", user.Username, user.Id, clan.Name, clan.Id)
	return nil
}

// GetClan Retrieves data about an individual clan
// GET /v2/clan/:id
func GetClan(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	clan, err := db.GetClanById(id)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("Clan")
	default:
		return APIErrorServerError("Error while fetching clan", err)
	}

	c.JSON(http.StatusOK, struct {
		Clan *db.Clan `json:"clan"`
	}{
		Clan: clan,
	})

	return nil
}

// UpdateClan Updates data about a clan
// Endpoint: PATCH /v2/clan/:id
func UpdateClan(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	clan, apiErr := getClanAndCheckOwnership(user, id)

	if apiErr != nil {
		return apiErr
	}

	body := struct {
		Name         *string `form:"name" json:"name"`
		Tag          *string `form:"tag" json:"tag"`
		FavoriteMode *uint8  `form:"favorite_mode" json:"favorite_mode"`
		AboutMe      *string `form:"about_me" json:"about_me"`
		AccentColor  *string `form:"accent_color" json:"accent_color"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if !clan.IsCustomizable {
		return APIErrorForbidden("Your clan must be customizable for you to access this endpoint")
	}

	if body.Name != nil {
		if !db.IsValidClanName(*body.Name) {
			return APIErrorBadRequest(errClanNameInvalid)
		}

		exists, err := db.DoesClanExistByName(*body.Name)

		if err != nil {
			return APIErrorServerError("Error checking if clan exists by name", err)
		}

		if exists && strings.ToLower(clan.Name) != strings.ToLower(*body.Name) {
			return APIErrorBadRequest(errClanNameExists)
		}

		if err := clan.UpdateName(*body.Name); err != nil {
			return APIErrorServerError("Error updating clan name", err)
		}
	}

	if body.Tag != nil {
		if !db.IsValidClanTag(*body.Tag) {
			return APIErrorBadRequest(errClanTagInvalid)
		}

		tagExists, existingClan, err := db.DoesClanExistByTag(*body.Tag)

		if err != nil {
			return APIErrorServerError("Error checking if clan exists by tag", err)
		}

		if tagExists && existingClan.Id != clan.Id {
			return APIErrorBadRequest(errClanTagExists)
		}

		if err := clan.UpdateTag(*body.Tag); err != nil {
			return APIErrorServerError("Error updating clan tag", err)
		}
	}

	if body.FavoriteMode != nil {
		if *body.FavoriteMode < 1 || *body.FavoriteMode > 2 {
			return APIErrorBadRequest(errClanFavoriteModeInvalid)
		}

		clan.FavoriteMode = *body.FavoriteMode

		if err := clan.UpdateFavoriteMode(enums.GameMode(*body.FavoriteMode)); err != nil {
			return APIErrorServerError("Error updating clan favorite mode", err)
		}
	}

	if body.AboutMe != nil {
		if len(*body.AboutMe) > 2000 {
			return APIErrorBadRequest(errClanAboutMeInvalid)
		}

		sanitized := stringutil.SanitizeHTML(*body.AboutMe)
		clan.AboutMe = &sanitized

		if err := clan.UpdateAboutMe(*clan.AboutMe); err != nil {
			return APIErrorServerError("Error updating clan favorite mode", err)
		}
	}

	if body.AccentColor != nil {
		if !stringutil.IsValidHexCode(*body.AccentColor) {
			return APIErrorBadRequest(errClanAccentColorInvalid)
		}

		clan.AccentColor = body.AccentColor

		if err := clan.UpdateAccentColor(*body.AccentColor); err != nil {
			return APIErrorServerError("Error updating clan accent color", err)
		}
	}

	if err := clan.UpdateLastUpdated(); err != nil {
		return APIErrorServerError("Error updating last_updated for clan", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your clan has been successfully updated."})
	return nil
}

// DeleteClan Deletes an individual clan
// Endpoint: DELETE /v2/clan/:id
func DeleteClan(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	clan, apiErr := getClanAndCheckOwnership(user, id)

	if apiErr != nil {
		return apiErr
	}

	err = db.DeleteClan(clan.Id)

	if err != nil {
		return APIErrorServerError("Error while deleting clan", err)
	}

	if err := db.RemoveClanFromLeaderboards(clan.Id); err != nil {
		return APIErrorServerError("Error removing clan from leaderboard", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your clan has been successfully deleted."})
	logrus.Debugf("%v (#%v) has deleted the clan: `%v` (#%v).", user.Username, user.Id, clan.Name, clan.Id)
	return nil
}

// Selects a clan from the database with clanId and checks if the user is the owner.
func getClanAndCheckOwnership(user *db.User, clanId int) (*db.Clan, *APIError) {
	clan, err := db.GetClanById(clanId)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return nil, APIErrorNotFound("Clan")
	default:
		return nil, APIErrorServerError("Error while retrieving clan from db", err)
	}

	if clan.OwnerId != user.Id || clan.Id != *user.ClanId {
		return nil, &APIError{Status: http.StatusForbidden, Message: "You are not the owner of the clan."}
	}

	return clan, nil
}
