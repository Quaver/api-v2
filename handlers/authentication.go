package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

// AuthenticateUser Authenticates a user from an incoming HTTP request
func AuthenticateUser(c *gin.Context) *db.User {
	jwt := c.GetHeader("Authorization")
	inGameToken := c.GetHeader("auth")

	var user *db.User
	var err error

	if jwt != "" {
		user, err = authenticateJWT(jwt)
	} else if inGameToken != "" {
		user, err = authenticateInGame(inGameToken)
	} else {
		ReturnError(c, http.StatusUnauthorized, "You must provide an `Authorization` or `auth` header to access this resource.")
		return nil
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("An error occurred while authenticating user: %v", err)
		Return500(c)
		return nil
	}

	if user == nil {
		ReturnError(c, http.StatusUnauthorized, "The authentication token you have provided was not valid.")
		return nil
	}

	return user
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
