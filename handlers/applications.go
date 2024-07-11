package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/stringutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"strconv"
	"time"
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

// CreateNewApplication Creates a new application
// Endpoint: POST /v2/developers/applications
func CreateNewApplication(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Name        string `form:"name" json:"name" binding:"required"`
		RedirectURL string `form:"redirect_url" json:"redirect_url" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if len(body.Name) > 50 {
		return APIErrorBadRequest("The name of your application must not exceed 50 characters.")
	}

	_, err := url.ParseRequestURI(body.RedirectURL)

	if err != nil {
		return APIErrorBadRequest("Your redirect URL is not valid.")
	}

	applications, err := db.GetUserActiveApplications(user.Id)

	if err != nil {
		return APIErrorServerError("Error fetching user active applications", err)
	}

	if len(applications) >= 10 {
		return APIErrorForbidden("You have already created the maximum amount of applications.")
	}

	clientId, err := stringutil.GenerateToken(16)

	if err != nil {
		return APIErrorServerError("Error generating client id", err)
	}

	clientSecret, err := stringutil.GenerateToken(32)

	if err != nil {
		return APIErrorServerError("Error generating client secret", err)
	}

	newApp := &db.Application{
		UserId:       user.Id,
		Name:         body.Name,
		RedirectURL:  body.RedirectURL,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Timestamp:    time.Now().UnixMilli(),
		Active:       true,
	}

	if err := db.SQL.Create(&newApp).Error; err != nil {
		return APIErrorServerError("Error inserting application into db", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Your newApp has been successfully created.",
		"application": newApp,
	})

	return nil
}

// ResetApplicationSecret Resets the secret of an application
// Endpoint: POST /v2/developers/applications/:id/secret
func ResetApplicationSecret(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	application, apiErr := getUserApplication(c, user)

	if apiErr != nil {
		return apiErr
	}

	secret, err := stringutil.GenerateToken(32)

	if err != nil {
		return APIErrorServerError("Error generating random string", err)
	}

	if err := application.SetClientSecret(secret); err != nil {
		return APIErrorServerError("Error setting application secret", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "You have successfully reset your secret.",
		"secret":  application.ClientSecret,
	})

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
