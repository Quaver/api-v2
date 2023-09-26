package handlers

import (
	"github.com/gin-gonic/gin"
)

// HandleInviteUserToClan InviteUserToClan Invites a user to their clan
// POST /clan/invite
func HandleInviteUserToClan(c *gin.Context) {
	if err := inviteUserToClan(c); err != nil {
		handleAPIError(c, err)
		return
	}
}

func inviteUserToClan(c *gin.Context) *APIError {
	return nil
	/*	user, apiErr := authenticateUser(c)

		if apiErr != nil {
			return apiErr
		}

		if user.ClanId == nil {
		//	return
		}

		_, err := getClanAndCheckOwnership(c, user, *user.ClanId)

		if err != nil {
		//	return
		}

		body := struct {
			UserId int `form:"user_id" json:"user_id" binding:"required"`
		}{}

		if err := c.ShouldBind(&body); err != nil {
			//	Return400(c)
		//	return
		}

		if user.Id == body.UserId {
			ReturnError(c, http.StatusBadRequest, "You cannot invite yourself to the clan.")
		//	return
		}

		invitee, err := db.GetUserById(body.UserId)

		switch err {
		case nil:
			break
		case gorm.ErrRecordNotFound:
			ReturnError(c, http.StatusBadRequest, "That user does not exist.")
			return
		default:
			logrus.Error("Error getting user in db: ", err)
			//	Return500(c)
			return
		}

		if invitee.ClanId != nil {
			ReturnError(c, http.StatusBadRequest, "You cannot invite that user because they are already in a clan.")
			return
		}

		// TODO: Check for pending invitation
		// TODO: Invite the user
		c.JSON(http.StatusOK, "INVITED USER TO CLAN")*/
}
