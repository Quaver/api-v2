package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

// authenticateUser Authenticates a user from an incoming HTTP request
func authenticateUser(c *gin.Context) (*db.User, *APIError) {
	jwt := c.GetHeader("Authorization")
	inGameToken := c.GetHeader("auth")

	var user *db.User
	var err error

	if jwt != "" {
		user, err = authenticateJWT(jwt)
	} else if inGameToken != "" {
		user, err = authenticateInGame(inGameToken)
	} else {
		return nil, &APIError{Status: http.StatusUnauthorized, Message: "You must provide a valid `Authorization` or `auth` header."}
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, APIErrorServerError("Error occurred while authenticating user", err)
	}

	if user == nil {
		return nil, &APIError{Status: http.StatusUnauthorized, Message: "The authentication token you have provided was not valid"}
	}

	if !user.Allowed {
		return nil, APIErrorForbidden("You are banned")
	}

	return user, nil
}

// Authenticates a user by their JWT token.
// Header - {Authorization: 'Bearer Token`}
func authenticateJWT(token string) (*db.User, error) {
	logrus.Warn("Please implement db.authenticateJWT().")

	user, err := db.GetUserById(1)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticates a user by their in-game token.
// Header - {Auth:Token}
func authenticateInGame(token string) (*db.User, error) {
	logrus.Warn("Please implement db.authenticateInGame().")
	return nil, nil
}
