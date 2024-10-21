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
	errClanInviteDoesntBelong string = "This clan invite does not belong to you."
)

// InviteUserToClan Invites a user to the clan
// Endpoint: POST /v2/clan/invite
func InviteUserToClan(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
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

	if apiErr := checkMaxClanMembers(clan.Id); apiErr != nil {
		return apiErr
	}

	invitingUser, err := db.GetUserById(body.UserId)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorBadRequest("That user does not exist.")
	default:
		return APIErrorServerError("Error retrieving user in db", err)
	}

	if invitingUser.ClanId != nil {
		return APIErrorBadRequest("You cannot invite that user because they are already in a clan.")
	}

	pendingInvite, err := db.GetPendingClanInvite(clan.Id, invitingUser.Id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error while retrieving pending clan invite", err)
	}

	if pendingInvite != nil {
		return APIErrorBadRequest("That user already has a pending invite to the clan.")
	}

	clanInvite, err := db.InviteUserToClan(clan.Id, invitingUser.Id)

	if err != nil {
		return APIErrorServerError("Error inserting invite into database", err)
	}

	if err := db.NewClanInviteNotification(clan, clanInvite).Insert(); err != nil {
		return APIErrorServerError("Error inserting clan invite notification", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "You have successfully invited the user to the clan.",
		"invite":  clanInvite,
	})

	logrus.Debugf("%v (#%v) has invited %v (#%v) to clan: %v (#%v)", user.Username, user.Id, invitingUser.Username, invitingUser.Id, clan.Name, clan.Id)
	return nil
}

// GetClanInvite Retrieves a clan invite by id
// Endpoint: GET /v2/clan/invite/:id
func GetClanInvite(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	invite, apiErr := parseAndGetClanInvite(c, user)

	if apiErr != nil {
		return apiErr
	}

	c.JSON(http.StatusOK, gin.H{"clan_invite": invite})
	return nil
}

// AcceptClanInvite Accepts a clan invitation and joins the clan
// Endpoint: POST /v2/clan/invite/:id/accept
func AcceptClanInvite(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !user.CanJoinClan() {
		return APIErrorBadRequest(errClanUserCantJoin)
	}

	invite, apiErr := parseAndGetClanInvite(c, user)

	if apiErr != nil {
		return apiErr
	}

	if apiErr := checkMaxClanMembers(invite.ClanId); apiErr != nil {
		return apiErr
	}

	if err := db.UpdateUserClan(user.Id, invite.ClanId); err != nil {
		return APIErrorServerError("Error updating user clan", err)
	}

	if err := db.DeleteUserClanInvites(user.Id); err != nil {
		return APIErrorServerError("Error deleting user clan invites", err)
	}

	if err := db.DeleteUserClanInviteNotifications(user.Id); err != nil {
		return APIErrorServerError("Error deleting user clan invite notifications", err)
	}

	if err := db.NewClanActivity(invite.ClanId, db.ClanActivityUserJoined, user.Id).Insert(); err != nil {
		return APIErrorServerError("Error inserting clan activity", err)
	}

	logrus.Debugf("%v (#%v) has joined the clan #%v", user.Username, user.Id, invite.ClanId)
	c.JSON(http.StatusOK, gin.H{"message": "You have successfully joined the clan."})
	return nil
}

// DeclineClanInvite Declines a clan invite and deletes it
// Endpoint: POST /v2/clan/invite/:id/decline
func DeclineClanInvite(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	invite, err := db.GetClanInviteById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving clan invite from db", err)
	}

	if invite == nil {
		return APIErrorNotFound("Invite")
	}

	clan, err := db.GetClanById(invite.ClanId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving clan from db", err)
	}

	if clan == nil {
		return APIErrorNotFound("Clan")
	}

	if invite.UserId != user.Id && user.Id != clan.OwnerId {
		return APIErrorForbidden(errClanInviteDoesntBelong)
	}

	if err := db.DeleteClanInviteById(invite.Id); err != nil {
		return APIErrorServerError("Error deleting clan invite", err)
	}

	logrus.Debugf("%v (#%v) has declined clan invite #%v", user.Username, user.Id, invite.Id)
	c.JSON(http.StatusOK, gin.H{"message": "You have successfully declined the clan invite."})
	return nil
}

// GetClanPendingInvites Returns a clan's pending invites
// Endpoint: GET /v2/clan/invites
func GetClanPendingInvites(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if user.ClanId == nil {
		return APIErrorForbidden("You are not currently in a clan.")
	}

	clan, apiErr := getClanAndCheckOwnership(user, *user.ClanId)

	if apiErr != nil {
		return apiErr
	}

	invites, err := db.GetPendingClanInvites(clan.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving clan invites from db", err)
	}

	c.JSON(http.StatusOK, gin.H{"pending_invites": invites})
	return nil
}

// GetUserPendingClanInvites Retrieves all pending clan invites for the user
// Endpoint: GET /v2/clan/invites
func GetUserPendingClanInvites(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	invites, err := db.GetUserClanInvites(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving user clan invites", err)
	}

	c.JSON(http.StatusOK, gin.H{"clan_invites": invites})
	return nil
}

// Expects an id param on gin.Context for a clan invitation. Then parses it,
// and checks if the invite belongs to the user
func parseAndGetClanInvite(c *gin.Context, user *db.User) (*db.ClanInvite, *APIError) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return nil, APIErrorBadRequest("Invalid id")
	}

	invite, err := db.GetClanInviteById(id)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return nil, APIErrorNotFound("Clan invite")
	default:
		return nil, APIErrorServerError("Error retrieving clan invite from the database", err)
	}

	if invite.UserId != user.Id {
		return nil, APIErrorForbidden(errClanInviteDoesntBelong)
	}

	return invite, nil
}

func checkMaxClanMembers(clanId int) *APIError {
	memberCount, err := db.GetClanMemberCount(clanId)

	if err != nil {
		return APIErrorServerError("Error retrieving clan member count", err)
	}

	if memberCount >= 100 {
		return APIErrorBadRequest("The clan you are trying to join is already full.")
	}

	return nil
}
