package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

const (
	errClanUserCantJoin string = "You must not be in clan, and wait at least 1 day after leaving your previous clan to create a new one."
	errClanNameInvalid  string = "Your clan `name` must be between 3 and 30 characters and must contain only letters or numbers.."
	errClanTagInvalid   string = "Your clan `tag` must be between 1 and 4 characters and must contain only letters or numbers."
	errClanNameExists   string = "A clan with that name already exists. Please choose a different name."
)

// GetClans Retrieves basic info / leaderboard data about clans
// Endpoint: GET /v2/clans?page=1
func GetClans(c *gin.Context) {
}

// CreateClan Creates a new clan if the user is eligible to.
// Endpoint: POST /v2/clan
func CreateClan(c *gin.Context) {
	user := AuthenticateUser(c)

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

	existingClan, err := db.GetClanByName(body.Name)

	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("Error retrieving clan by name `%v`: %v", body.Name, err)
		Return500(c)
		return
	}

	if existingClan != nil {
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
}

// DeleteClan Deletes an individual clan
// Endpoint: DELETE /v2/clan/:id
func DeleteClan(c *gin.Context) {
}
