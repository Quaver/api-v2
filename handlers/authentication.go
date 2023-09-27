package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

// AuthenticateUser Middleware authentication function
func AuthenticateUser(c *gin.Context) {
	user, apiErr := authenticateUser(c)

	if apiErr != nil {
		CreateHandler(func(ctx *gin.Context) *APIError {
			return apiErr
		})(c)

		c.Abort()
		return
	}

	c.Set("user", user)
	c.Next()
}

// Returns an authenticated user from a context
func getAuthedUser(c *gin.Context) *db.User {
	user, exists := c.Get("user")

	if !exists {
		return nil
	}

	return user.(*db.User)
}

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
