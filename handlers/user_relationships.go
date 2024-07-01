package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// AddFriend Adds a friend to the logged-in user's friends list
// Endpoint: POST /v2/user/:id/relationship/add
func AddFriend(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if user.Id == id {
		return APIErrorBadRequest("You cannot add yourself.")
	}

	relationship, err := db.GetUserRelationship(user.Id, id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving user relationship", err)
	}

	if relationship != nil {
		return APIErrorBadRequest("You are already friends with that user.")
	}

	err = db.AddFriend(user.Id, id)

	if err != nil {
		return APIErrorServerError("Error adding friend", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully added that user to your friend's list!"})
	return nil
}

// RemoveFriend Removes a friend from the logged-in users friend's list
// Endpoint: POST /v2/user/:id/relationship/delete
func RemoveFriend(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	err = db.RemoveFriend(user.Id, id)

	if err != nil {
		return APIErrorServerError("Error removing friend", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully removed that user from your friend's list!"})
	return nil
}

// GetFriendsList Retrieves the logged-in user's friend's list
// Endpoint: GET /v2/user/relationship/friends
func GetFriendsList(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	friends, err := db.GetUserFriends(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving friends list", err)
	}

	c.JSON(http.StatusOK, gin.H{"friends": friends})
	return nil
}
