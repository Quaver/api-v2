package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	errClanUserCantJoin        string = "You must not already be in a clan to perform this action."
	errClanNameInvalid         string = "Your clan `name` must be between 3 and 30 characters and must contain only letters or numbers."
	errClanTagInvalid          string = "Your clan `tag` must be between 1 and 4 characters and must contain only letters or numbers."
	errClanAboutMeInvalid      string = "Your clan `about_me` must be between 0 and 2000 characters."
	errClanFavoriteModeInvalid string = "Your clan `favorite_mode` must be a valid mode id."
	errClanNameExists          string = "A clan with that name already exists. Please choose a different name."
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

	exists, err := db.DoesClanExistByName(body.Name)

	if err != nil {
		return APIErrorServerError("Error checking if clan exists by name", err)
	}

	if exists {
		return APIErrorBadRequest(errClanNameExists)
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

	if err := db.NewClanActivity(clan.Id, db.ClanActivityCreated, user.Id).InsertClanActivity(); err != nil {
		return APIErrorServerError("Error inserting clan activity", err)
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
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
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

		clan.Name = *body.Name
		clan.LastNameChangeTime = time.Now().UnixMilli()
	}

	if body.Tag != nil {
		if !db.IsValidClanTag(*body.Tag) {
			return APIErrorBadRequest(errClanTagInvalid)
		}

		clan.Tag = *body.Tag
	}

	if body.FavoriteMode != nil {
		if *body.FavoriteMode < 1 || *body.FavoriteMode > 2 {
			return APIErrorBadRequest(errClanFavoriteModeInvalid)
		}

		clan.FavoriteMode = *body.FavoriteMode
	}

	if body.AboutMe != nil {
		if len(*body.AboutMe) > 2000 {
			return APIErrorBadRequest(errClanAboutMeInvalid)
		}

		clan.AboutMe = body.AboutMe
	}

	if result := db.SQL.Save(clan); result.Error != nil {
		return APIErrorServerError("Error updating clan in the database", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your clan has been successfully updated."})
	logrus.Debugf("%v (#%v) has updated the clan: `%v` (#%v).", user.Username, user.Id, clan.Name, clan.Id)
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
