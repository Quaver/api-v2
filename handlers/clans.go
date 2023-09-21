package handlers

import (
	"errors"
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
	errClanUserCantJoin        string = "You must not be in clan, and wait at least 1 day after leaving your previous clan to create a new one."
	errClanNameInvalid         string = "Your clan `name` must be between 3 and 30 characters and must contain only letters or numbers."
	errClanTagInvalid          string = "Your clan `tag` must be between 1 and 4 characters and must contain only letters or numbers."
	errClanAboutMeInvalid      string = "Your clan `about_me` must be between 0 and 2000 characters."
	errClanFavoriteModeInvalid string = "Your clan `favorite_mode` must be a valid mode id."
	errClanNameExists          string = "A clan with that name already exists. Please choose a different name."
)

// GetClans Retrieves basic info / leaderboard data about clans
// Endpoint: GET /v2/clans?page=1
func GetClans(c *gin.Context) {
}

// CreateClan Creates a new clan if the user is eligible to.
// Endpoint: POST /v2/clan
func CreateClan(c *gin.Context) {
	user := authenticateUser(c)

	if user == nil {
		return
	}

	if !user.CanJoinClan() {
		ReturnError(c, http.StatusBadRequest, errClanUserCantJoin)
		return
	}

	body := struct {
		Name string `form:"name" json:"name" binding:"required"`
		Tag  string `form:"tag" json:"tag" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		Return400(c)
		return
	}

	if !db.IsValidClanName(body.Name) {
		ReturnError(c, http.StatusBadRequest, errClanNameInvalid)
		return
	}

	if !db.IsValidClanTag(body.Tag) {
		ReturnError(c, http.StatusBadRequest, errClanTagInvalid)
		return
	}

	exists, err := db.DoesClanExistByName(body.Name)

	if err != nil {
		logrus.Errorf("Error checking if clan exists by name - %v", err)
		Return500(c)
		return
	}

	if exists {
		ReturnError(c, http.StatusBadRequest, errClanNameExists)
		return
	}

	clan := db.Clan{
		OwnerId: user.Id,
		Name:    body.Name,
		Tag:     body.Tag,
	}

	if err := clan.Insert(); err != nil {
		logrus.Error("Error inserting clan into database: ", err)
		Return500(c)
		return
	}

	err = db.UpdateUserClan(user.Id, clan.Id)

	if err != nil {
		logrus.Error("Error while updating user clan: ", err)
		Return500(c)
		return
	}

	c.JSON(http.StatusOK, struct {
		Message string   `json:"message"`
		Clan    *db.Clan `json:"clan"`
	}{
		Message: "Your clan has been successfully created.",
		Clan:    &clan,
	})

	logrus.Debugf("%v (#%v) has created the clan `%v` (#%v).", user.Username, user.Id, clan.Name, clan.Id)
}

// GetClan Retrieves data about an individual clan
// GET /v2/clan/:id
func GetClan(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		Return400(c)
		return
	}

	clan, err := db.GetClanById(id)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Clan not found"})
		return
	default:
		logrus.Errorf("Error while fetching clan: `%v` - %v", id, err)
		Return500(c)
		return
	}

	c.JSON(http.StatusOK, struct {
		Clan *db.Clan `json:"clan"`
	}{
		Clan: clan,
	})
}

// UpdateClan Updates data about a clan
// Endpoint: PATCH /v2/clan/:id
func UpdateClan(c *gin.Context) {
	user := authenticateUser(c)

	if user == nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		Return400(c)
		return
	}

	clan, err := getClanAndCheckOwnership(c, user, id)

	if err != nil {
		return
	}

	body := struct {
		Name         *string `form:"name" json:"name"`
		Tag          *string `form:"tag" json:"tag"`
		FavoriteMode *uint8  `form:"favorite_mode" json:"favorite_mode"`
		AboutMe      *string `form:"about_me" json:"about_me"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		Return400(c)
		return
	}

	// Update Name
	if body.Name != nil {
		if !db.IsValidClanName(*body.Name) {
			ReturnError(c, http.StatusBadRequest, errClanNameInvalid)
			return
		}

		exists, err := db.DoesClanExistByName(*body.Name)

		if err != nil {
			logrus.Errorf("Error checking if clan exists by name - %v", err)
			Return500(c)
			return
		}

		if exists && strings.ToLower(clan.Name) != strings.ToLower(*body.Name) {
			ReturnError(c, http.StatusBadRequest, errClanNameExists)
			return
		}

		clan.Name = *body.Name
		clan.LastNameChangeTime = time.Now().UnixMilli()
	}

	// Update Tag
	if body.Tag != nil {
		if !db.IsValidClanTag(*body.Tag) {
			ReturnError(c, http.StatusBadRequest, errClanTagInvalid)
			return
		}

		clan.Tag = *body.Tag
	}

	// Update Favorite Mode
	if body.FavoriteMode != nil {
		if *body.FavoriteMode < 1 || *body.FavoriteMode > 2 {
			ReturnError(c, http.StatusBadRequest, errClanFavoriteModeInvalid)
			return
		}

		clan.FavoriteMode = *body.FavoriteMode
	}

	// Update About Me
	if body.AboutMe != nil {
		if len(*body.AboutMe) > 2000 {
			ReturnError(c, http.StatusBadRequest, errClanAboutMeInvalid)
			return
		}

		clan.AboutMe = body.AboutMe
	}

	if result := db.SQL.Save(clan); result.Error != nil {
		logrus.Errorf("Error updating clan: %v (#%v) in the database - %v", clan.Name, clan.Id, result.Error)
		Return500(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your clan has been successfully updated."})
	logrus.Debugf("%v (#%v) has updated the clan: `%v` (#%v).", user.Username, user.Id, clan.Name, clan.Id)
}

// DeleteClan Deletes an individual clan
// Endpoint: DELETE /v2/clan/:id
func DeleteClan(c *gin.Context) {
	user := authenticateUser(c)

	if user == nil {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		Return400(c)
		return
	}

	clan, err := getClanAndCheckOwnership(c, user, id)

	if err != nil {
		return
	}

	c.JSON(http.StatusOK, clan)
	logrus.Debugf("%v (#%v) has deleted the clan: `%v` (#%v).", user.Username, user.Id, clan.Name, clan.Id)
}

// Selects a clan from the database with clanId and checks if the user is the owner.
func getClanAndCheckOwnership(c *gin.Context, user *db.User, clanId int) (*db.Clan, error) {
	clan, err := db.GetClanById(clanId)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Clan not found"})
		return nil, err
	default:
		logrus.Errorf("Erorr while retrieving clan from db - %v", err)
		Return500(c)
		return nil, err
	}

	if clan.OwnerId != user.Id || clan.Id != *user.ClanId {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the owner of the clan."})
		return nil, errors.New("user is not clan owner")
	}

	return clan, nil
}
