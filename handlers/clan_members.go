package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// Struct that is returned from checkClanOwnerAndGetTargetUser.
// This is used for any endpoints where the clan owner needs to perform an action on a target user.
type targetClanMember struct {
	ClanOwner  *db.User
	TargetUser *db.User
	Clan       *db.Clan
	Error      *APIError
}

// GetClanMembers Retrieves a list of members in the clan
// GET /v2/clan/:id/members
func GetClanMembers(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	_, err = db.GetClanById(id)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("Clan")
	default:
		return APIErrorServerError("Error while fetching clan", err)
	}

	members, err := db.GetUsersInClan(id)

	if err != nil {
		return APIErrorServerError("Error retrieving users in clan", err)
	}

	c.JSON(http.StatusOK, struct {
		ClanMembers []*db.User `json:"clan_members"`
	}{
		ClanMembers: members,
	})

	return nil
}

// LeaveClan Leaves the user's current clan
// Endpoint: POST /v2/clan/leave
func LeaveClan(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if user.ClanId == nil {
		return APIErrorBadRequest("You are currently not in a clan.")
	}

	clan, err := db.GetClanById(*user.ClanId)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorNotFound("Clan does not exist. Please report this to a developer.")
	default:
		return APIErrorServerError("Error retrieving clan from db", err)
	}

	if clan.OwnerId == user.Id {
		return APIErrorBadRequest("You cannot leave the clan while you are the owner.")
	}

	if err := db.UpdateUserClan(user.Id); err != nil {
		return APIErrorServerError("Error updating user clan", err)
	}

	if err := db.NewClanActivity(clan.Id, db.ClanActivityUserLeft, user.Id).Insert(); err != nil {
		return APIErrorServerError("Error inserting clan activity", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully left the clan."})
	logrus.Debugf("%v (#%v) has left the clan: %v (#%v)", user.Username, user.Id, clan.Name, clan.Id)
	return nil
}

// TransferClanOwnership Transfers ownership of the clan to another member.
// Endpoint: POST /v2/clan/transfer/:user_id
func TransferClanOwnership(c *gin.Context) *APIError {
	target := checkClanOwnerAndGetTargetUser(c)

	if target.Error != nil {
		return target.Error
	}

	target.Clan.OwnerId = target.TargetUser.Id

	if result := db.SQL.Save(target.Clan); result.Error != nil {
		return APIErrorServerError("Error updating clan in the database", result.Error)
	}

	if err := db.NewClanActivity(target.Clan.Id, db.ClanActivityOwnershipTransferred, target.TargetUser.Id).Insert(); err != nil {
		return APIErrorServerError("Error inserting clan activity", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully transferred ownership of the clan."})
	return nil
}

// KickClanMember Kicks a member from the clan
// POST /v2/clan/kick/:user_id
func KickClanMember(c *gin.Context) *APIError {
	target := checkClanOwnerAndGetTargetUser(c)

	if target.Error != nil {
		return target.Error
	}

	if err := db.UpdateUserClan(target.TargetUser.Id); err != nil {
		return APIErrorServerError("Error updating user clan", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully kicked that user from the clan."})
	return nil
}

// Helper function which checks clan ownership and a target user to perform an action on.
// This is useful for things like kicking a user, checking ownership, or where the clan owner needs to perform an action on a member.
// This will return (the clan owner, the target user, the clan, and any api error)
func checkClanOwnerAndGetTargetUser(c *gin.Context) *targetClanMember {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if user.ClanId == nil {
		return &targetClanMember{Error: APIErrorBadRequest("You are currently not in a clan")}
	}

	clan, apiErr := getClanAndCheckOwnership(user, *user.ClanId)

	if apiErr != nil {
		return &targetClanMember{Error: apiErr}
	}

	newOwnerId, err := strconv.Atoi(c.Param("user_id"))

	if err != nil {
		return &targetClanMember{Error: APIErrorBadRequest("Invalid user_id")}
	}

	if newOwnerId == user.Id {
		return &targetClanMember{Error: APIErrorBadRequest("You cannot perform this action on yourself")}
	}

	targetUser, err := db.GetUserById(newOwnerId)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return &targetClanMember{Error: APIErrorBadRequest("That user does not exist")}
	default:
		return &targetClanMember{Error: APIErrorServerError("Error getting user by id", err)}
	}

	if targetUser.ClanId == nil || *targetUser.ClanId != clan.Id {
		return &targetClanMember{Error: APIErrorBadRequest("That user must be in your clan to perform an action on them.")}
	}

	return &targetClanMember{
		ClanOwner:  user,
		TargetUser: targetUser,
		Clan:       clan,
		Error:      nil,
	}
}
