package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func InviteUserToClan(c *gin.Context) *APIError {
	user, apiErr := authenticateUser(c)

	if apiErr != nil {
		return apiErr
	}

	if user.ClanId == nil {
		return APIErrorForbidden("You are not currently in a clan.")
	}

	_, apiErr = getClanAndCheckOwnership(user, *user.ClanId)

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

	// TODO: Check for pending invitation
	// TODO: Invite the user
	c.JSON(http.StatusOK, "INVITED USER TO CLAN")
	return nil
}
