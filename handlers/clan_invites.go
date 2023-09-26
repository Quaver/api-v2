package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

// InviteUserToClan Invites a user to the clan
// Endpoint: POST /v2/clan/invite
func InviteUserToClan(c *gin.Context) *APIError {
	user, apiErr := authenticateUser(c)

	if apiErr != nil {
		return apiErr
	}

	if user.ClanId == nil {
		return APIErrorForbidden("You are not currently in a clan.")
	}

	clan, apiErr := getClanAndCheckOwnership(user, *user.ClanId)

	if apiErr != nil {
		return apiErr
	}

	body := struct {
		UserId int `form:"user_id" json:"user_id" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if user.Id == body.UserId {
		return APIErrorBadRequest("You cannot invite yourself to the clan.")
	}

	invitee, err := db.GetUserById(body.UserId)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorBadRequest("That user does not exist.")
	default:
		return APIErrorServerError("Error retrieving user in db", err)
	}

	if invitee.ClanId != nil {
		return APIErrorBadRequest("You cannot invite that user because they are already in a clan.")
	}

	pendingInvite, err := db.GetPendingClanInvite(clan.Id, invitee.Id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error while retrieving pending clan invite", err)
	}

	if pendingInvite != nil {
		return APIErrorBadRequest("That user already has a pending invite to the clan.")
	}

	clanInvite, err := db.InviteUserToClan(clan.Id, invitee.Id)

	if err != nil {
		return APIErrorServerError("Error inserting invite into database", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "You have successfully invited the user to the clan.",
		"invite":  clanInvite,
	})

	logrus.Debugf("%v (#%v) has invited %v (#%v) to clan: %v (#%v)", user.Username, user.Id, invitee.Username, invitee.Id, clan.Name, clan.Id)
	return nil
}
