package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// GetRankingQueueComments Returns all of the comments for a mapset in the ranking queue
// Endpoint: GET /v2/ranking/queue/:id/comments
func GetRankingQueueComments(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	comments, err := db.GetRankingQueueComments(id)

	if err != nil {
		return APIErrorServerError("Error getting ranking queue comments", err)
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
	return nil
}

// AddRankingQueueComment Inserts a ranking queue comment to the database
// Endpoint: POST /v2/ranking/queue/:id/comment
func AddRankingQueueComment(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	body := struct {
		Comment string `form:"comment" json:"comment"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if len(body.Comment) == 0 || len(body.Comment) > 5000 {
		return APIErrorBadRequest("Your comment must be between 1 and 5,000 characters")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	queueMapset, err := db.GetRankingQueueMapset(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving ranking queue mapset from db", err)
	}

	if queueMapset == nil {
		return APIErrorNotFound("Mapset")
	}

	if queueMapset.Mapset.CreatorID != user.Id && enums.HasUserGroup(user.UserGroups, enums.UserGroupRankingSupervisor) {
		return APIErrorForbidden("You do not have permission to comment on this mapset.")
	}

	comment := &db.MapsetRankingQueueComment{
		UserId:   user.Id,
		MapsetId: queueMapset.MapsetId,
		Comment:  body.Comment,
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting comment into DB", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your comment has been successfully added."})
	return nil
}

// EditRankingQueueComment Edits a ranking queue comment
// Endpoint: POST /v2/ranking/queue/comment/:id/edit
func EditRankingQueueComment(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	body := struct {
		Comment string `form:"comment" json:"comment"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if len(body.Comment) == 0 || len(body.Comment) > 5000 {
		return APIErrorBadRequest("Your comment must be between 1 and 5,000 characters")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	comment, err := db.GetRankingQueueComment(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving ranking queue comment", err)
	}

	if comment == nil {
		return APIErrorNotFound("Comment")
	}

	if comment.UserId != user.Id {
		return APIErrorForbidden("You are not the author of this comment.")
	}

	comment.Comment = body.Comment
	comment.DateLastUpdated = time.Now().UnixMilli()

	if result := db.SQL.Save(comment); result.Error != nil {
		return APIErrorServerError("Error updating ranking queue comment in the database", result.Error)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your comment has been successfully edited."})
	return nil
}
