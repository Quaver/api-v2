package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthenticateUser Authenticates a user from an incoming HTTP request
func AuthenticateUser(c *gin.Context) *db.User {
	jwt := c.GetHeader("Authorization")
	inGameToken := c.GetHeader("auth")

	var user *db.User

	if jwt != "" {
		user = authenticateJWT(jwt)
	} else if inGameToken != "" {
		user = authenticateInGame(inGameToken)
	} else {
		ReturnError(c, http.StatusUnauthorized, "You must provide an `Authorization` or `auth` header to access this resource.")
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
func authenticateJWT(token string) *db.User {
	return nil
}

// Authenticates a user by their in-game token.
// Header - {Auth:Token}
func authenticateInGame(token string) *db.User {
	return nil
}
