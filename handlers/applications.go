package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// GetUserApplications Returns a users active applications
// Endpoint: GET /v2/developers/applications
func GetUserApplications(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	applications, err := db.GetUserActiveApplications(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving applications from the db", err)
	}

	for _, app := range applications {
		app.ClientSecret = ""
	}

	c.JSON(http.StatusOK, gin.H{"applications": applications})
	return nil
}

// GetUserApplication Returns a single user application
// Endpoint: GET /v2/developers/applications/:id
func GetUserApplication(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	application, apiErr := getUserApplication(c, user)

	if apiErr != nil {
		return apiErr
	}

	application.ClientSecret = ""

	c.JSON(http.StatusOK, gin.H{"application": application})
	return nil
}

// DeleteUserApplication Deletes a user's application
// Endpoint: DELETE /v2/developers/applications/:id
func DeleteUserApplication(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	application, apiErr := getUserApplication(c, user)

	if apiErr != nil {
		return apiErr
	}

	if err := application.SetActiveStatus(false); err != nil {
		return APIErrorServerError("Error setting application active status", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully deleted your application."})
	return nil
}

// Gets a user's application and checks if they're the owner.
func getUserApplication(c *gin.Context, user *db.User) (*db.Application, *APIError) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return nil, APIErrorBadRequest("Invalid id")
	}

	application, err := db.GetApplicationById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, APIErrorServerError("Error retrieving application from db", err)
	}

	if application == nil {
		return nil, APIErrorNotFound("Application")
	}

	if application.UserId != user.Id {
		return nil, APIErrorForbidden("You are not the owner of this application.")
	}

	return application, nil
}