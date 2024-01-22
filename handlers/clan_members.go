package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

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

	err = db.UpdateUserClan(user.Id)

	if err != nil {
		return APIErrorServerError("Error updating user clan", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully left the clan."})
	logrus.Debugf("%v (#%v) has left the clan: %v (#%v)", user.Username, user.Id, clan.Name, clan.Id)
	return nil
}

// TransferClanOwnership Transfers ownership of the clan to another member.
// Endpoint: POST /v2/clan/transfer/:user_id
func TransferClanOwnership(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if user.ClanId == nil {
		return APIErrorBadRequest("You are currently not in a clan")
	}

	clan, apiErr := getClanAndCheckOwnership(user, *user.ClanId)

	if apiErr != nil {
		return apiErr
	}

	newOwnerId, err := strconv.Atoi(c.Param("user_id"))

	if err != nil {
		return APIErrorBadRequest("Invalid user_id")
	}

	if newOwnerId == user.Id {
		return APIErrorBadRequest("You cannot transfer ownership to yourself.")
	}

	newOwner, err := db.GetUserById(newOwnerId)

	switch err {
	case nil:
		break
	case gorm.ErrRecordNotFound:
		return APIErrorBadRequest("That user does not exist!")
	default:
		return APIErrorServerError("Error getting user by id", err)
	}

	if newOwner.ClanId == nil || *newOwner.ClanId != clan.Id {
		return APIErrorBadRequest("The user must be in your clan in order to transfer ownership to them.")
	}

	clan.OwnerId = newOwner.Id

	if result := db.SQL.Save(clan); result.Error != nil {
		return APIErrorServerError("Error updating clan in the database", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully transferred ownership of the clan."})
	return nil
}
