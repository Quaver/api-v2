package handlers

import (
	"net/http"
	"strconv"

	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/webhooks"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetRankingQueueComments Returns all of the comments for a mapset in the ranking queue
// Endpoint: GET /v2/ranking/queue/:id/comments
func GetRankingQueueComments(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	canViewPrivate := canUserAccessSupervisorRoute(c)
	comments, err := db.GetRankingQueueComments(id, canViewPrivate)

	if err != nil {
		return APIErrorServerError("Error getting ranking queue comments", err)
	}

	c.JSON(http.StatusOK, gin.H{"comments": prepareRankingQueueCommentsForResponse(comments, canViewPrivate)})
	return nil
}

func prepareRankingQueueCommentsForResponse(comments []*db.MapsetRankingQueueComment, canViewPrivate bool) []*db.MapsetRankingQueueComment {
	prepared := make([]*db.MapsetRankingQueueComment, 0, len(comments))

	for _, comment := range comments {
		if comment.IsPrivate && !canViewPrivate {
			continue
		}

		if comment.IsAnonymous {
			avatarUrl := webhooks.QuaverLogo
			redacted := *comment
			redacted.User = &db.User{
				Id:         db.QuaverBotId,
				Username:   "QuaverBot",
				UserGroups: enums.UserGroupBot,
				AvatarUrl:  &avatarUrl,
			}

			if canViewPrivate {
				redacted.AnonymousAuthor = comment.User
			} else {
				redacted.UserId = db.QuaverBotId
			}

			prepared = append(prepared, &redacted)
			continue
		}

		prepared = append(prepared, comment)
	}

	return prepared
}

// AddRankingQueueComment Inserts a ranking queue comment to the database
// Endpoint: POST /v2/ranking/queue/:id/comment
func AddRankingQueueComment(c *gin.Context) *APIError {
	return addRankingQueueComment(c, db.RankingQueueActionComment, false)
}

// AddPrivateRankingQueueComment inserts an attributed comment only ranking supervisors can retrieve.
// Endpoint: POST /v2/ranking/queue/:id/private-comment
func AddPrivateRankingQueueComment(c *gin.Context) *APIError {
	canUsePrivate := canUserAccessSupervisorRoute(c)
	if !canUsePrivate {
		return APIErrorForbidden("Only ranking supervisors can add a private ranking comment.")
	}

	return addRankingQueueComment(c, db.RankingQueueActionComment, true)
}

func addRankingQueueComment(c *gin.Context, action db.RankingQueueAction, isPrivate bool) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	body := struct {
		Comment  string         `form:"comment" json:"comment" binding:"required"`
		GameMode enums.GameMode `form:"game_mode" json:"game_mode"`
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

	if queueMapset.Mapset.CreatorID != user.Id && !enums.HasPrivilege(user.Privileges, enums.PrivilegeRankMapsets) {
		return APIErrorForbidden("You do not have permission to comment on this mapset.")
	}

	comment := &db.MapsetRankingQueueComment{
		UserId:     user.Id,
		MapsetId:   queueMapset.MapsetId,
		ActionType: action,
		Comment:    body.Comment,
		GameMode:   &body.GameMode,
		IsPrivate:  isPrivate,
		IsActive:   true,
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting comment into DB", err)
	}

	if action == db.RankingQueueActionComment {
		if err := db.NewMapsetActionNotification(queueMapset.Mapset, comment).Insert(); err != nil {
			return APIErrorServerError("Error inserting comment notification", err)
		}

		_ = webhooks.SendQueueWebhook(user, queueMapset.Mapset, db.RankingQueueActionComment)
	}

	message := "Your comment has been successfully added."
	if isPrivate {
		message = "Your private ranking comment has been successfully added."
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
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
		Comment string `form:"comment" json:"comment" binding:"required"`
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

	if err := comment.Edit(body.Comment); err != nil {
		return APIErrorServerError("Error updating ranking queue comment in the database", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your comment has been successfully edited."})
	return nil
}
