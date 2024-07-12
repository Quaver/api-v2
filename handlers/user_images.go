package handlers

import (
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/cache"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UploadUserProfileCover Handles the uploading of a user's profile cover
// Endpoint: POST /v2/user/profile/cover
func UploadUserProfileCover(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return APIErrorForbidden("You must be a donator to upload a profile cover.")
	}

	file, apiErr := validateUploadedImage(c)

	if apiErr != nil {
		return apiErr
	}

	_ = cache.RemoveCacheServerProfileCover(user.Id)
	
	err := azure.Client.UploadFile("profile-covers", fmt.Sprintf("%v.jpg", user.Id), file)

	if err != nil {
		return APIErrorServerError("Failed to upload file", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Your profile cover has been successfully updated!",
	})

	return nil
}
