package middleware

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

const (
	messageNoHeader = "You must provide a valid `Authorization` or `auth` header."
)

// RequireAuth Middleware authentication function
func RequireAuth(c *gin.Context) {
	user, apiErr := authenticateUser(c)

	if apiErr != nil {
		handlers.CreateHandler(func(ctx *gin.Context) *handlers.APIError {
			return apiErr
		})(c)

		c.Abort()
		return
	}

	c.Set("user", user)
	c.Next()
}

// AllowAuth Allows user authentication but does not require it. This middleware fails
// in the event that the user passes in an invalid token OR some other error
func AllowAuth(c *gin.Context) {
	user, apiErr := authenticateUser(c)

	if apiErr != nil && apiErr.Error != gorm.ErrRecordNotFound && apiErr.Message != messageNoHeader {
		handlers.CreateHandler(func(ctx *gin.Context) *handlers.APIError {
			return apiErr
		})(c)

		c.Abort()
		return
	}

	if user != nil {
		c.Set("user", user)
	}

	c.Next()
}

// authenticateUser Authenticates a user from an incoming HTTP request
func authenticateUser(c *gin.Context) (*db.User, *handlers.APIError) {
	jwt := c.GetHeader("Authorization")
	inGameToken := c.GetHeader("auth")

	var user *db.User
	var err error

	if jwt != "" {
		user, err = authenticateJWT(jwt)
	} else if inGameToken != "" {
		user, err = authenticateInGame(inGameToken)
	} else {
		return nil, &handlers.APIError{Status: http.StatusUnauthorized, Message: messageNoHeader}
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, handlers.APIErrorServerError("Error occurred while authenticating user", err)
	}

	if user == nil {
		return nil, &handlers.APIError{Status: http.StatusUnauthorized, Message: "The authentication token you have provided was not valid"}
	}

	if !user.Allowed {
		return nil, handlers.APIErrorForbidden("You are banned")
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
