package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

// AddNewGameBuild Adds a new game build to the database
// Endpoint: POST /v2/builds
func AddNewGameBuild(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	if !enums.HasPrivilege(user.Privileges, enums.PrivilegeManageBuilds) {
		return APIErrorForbidden("You do not have permission to manage game builds.")
	}

	body := struct {
		Version               string `form:"version" json:"version" binding:"required"`
		QuaverDll             string `form:"quaver_dll" json:"quaver_dll" binding:"required"`
		QuaverApiDll          string `form:"quaver_api_dll" json:"quaver_api_dll" binding:"required"`
		QuaverServerClientDll string `form:"quaver_server_client_dll" json:"quaver_server_client_dll" binding:"required"`
		QuaverServerCommonDll string `form:"quaver_server_common_dll" json:"quaver_server_common_dll" binding:"required"`
		QuaverSharedDll       string `form:"quaver_shared_dll" json:"quaver_shared_dll" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	build := &db.GameBuild{
		Version:               body.Version,
		QuaverDll:             body.QuaverDll,
		QuaverApiDll:          body.QuaverApiDll,
		QuaverServerClientDll: body.QuaverServerClientDll,
		QuaverServerCommonDll: body.QuaverServerCommonDll,
		QuaverSharedDll:       body.QuaverSharedDll,
	}

	if err := build.Insert(); err != nil {
		if err == gorm.ErrDuplicatedKey {
			return APIErrorBadRequest("You have already submitted a build with this version.")
		}

		return APIErrorServerError("Error inserting game build", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your build was successfully added to the database."})
	return nil
}
