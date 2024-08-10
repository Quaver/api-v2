package handlers

import (
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/webhooks"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type rankingQueueRequestData struct {
	MapsetId    int
	User        *db.User
	QueueMapset *db.RankingQueueMapset
	Comment     string
}

// Validates and returns common data used for ranking queue action requests
func validateRankingQueueRequest(c *gin.Context) (*rankingQueueRequestData, *APIError) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return nil, APIErrorBadRequest("You must provide a valid mapset id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil, APIErrorUnauthorized("Unauthorized")
	}

	body := struct {
		Comment string `form:"comment" json:"comment" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return nil, APIErrorBadRequest("Invalid request body")
	}

	if len(body.Comment) == 0 || len(body.Comment) > 5000 {
		return nil, APIErrorBadRequest("Your comment must be between 1 and 5,000 characters")
	}

	if !enums.HasPrivilege(user.Privileges, enums.PrivilegeRankMapsets) {
		return nil, APIErrorForbidden("You do not have permission to perform this action.")
	}

	queueMapset, err := db.GetRankingQueueMapset(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, APIErrorServerError("Error retrieving ranking queue mapset", err)
	}

	if queueMapset == nil {
		return nil, APIErrorNotFound("Mapset")
	}

	return &rankingQueueRequestData{
		MapsetId:    id,
		User:        user,
		QueueMapset: queueMapset,
		Comment:     body.Comment,
	}, nil
}

// VoteForRankingQueueMapset Adds a vote for a mapset in the ranking queue
// Endpoint: POST /v2/ranking/queue/:id/vote
func VoteForRankingQueueMapset(c *gin.Context) *APIError {
	data, apiErr := validateRankingQueueRequest(c)

	if apiErr != nil {
		return apiErr
	}

	queueMapset := data.QueueMapset

	if queueMapset.Mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("This mapset is already ranked.")
	}

	if queueMapset.Status != db.RankingQueuePending && queueMapset.Status != db.RankingQueueResolved {
		return APIErrorForbidden("This mapset must be either pending or resolved to be eligible to vote.")
	}

	if queueMapset.Mapset.CreatorID == data.User.Id {
		return APIErrorForbidden("You cannot vote for your own mapset.")
	}

	existingVotes, err := db.GetRankingQueueVotes(data.MapsetId)

	if err != nil {
		return APIErrorServerError("Error retrieving ranking queue votes", err)
	}

	for _, vote := range existingVotes {
		if vote.UserId == data.User.Id {
			return APIErrorForbidden("You have already voted for this mapset.")
		}

		if vote.User.IsTrialRankingSupervisor() && data.User.IsTrialRankingSupervisor() {
			return APIErrorForbidden("Two trial ranking supervisors cannot vote for the same mapset.")
		}
	}

	newVoteAction := &db.MapsetRankingQueueComment{
		UserId:     data.User.Id,
		User:       data.User,
		MapsetId:   data.MapsetId,
		ActionType: db.RankingQueueActionVote,
		IsActive:   true,
		Comment:    data.Comment,
	}

	existingVotes = append(existingVotes, newVoteAction)

	if err := newVoteAction.Insert(); err != nil {
		return APIErrorServerError("Error inserting new ranking queue vote", err)
	}

	if err := db.NewMapsetActionNotification(data.QueueMapset.Mapset, newVoteAction).Insert(); err != nil {
		return APIErrorServerError("Error inserting vote notification", err)
	}

	// Handle ranking the mapset
	if len(existingVotes) >= config.Instance.RankingQueue.VotesRequired {
		if err := queueMapset.UpdateStatus(db.RankingQueueRanked); err != nil {
			return APIErrorServerError("Error updating ranking queue status mapset status", err)
		}

		if err := db.RankMapset(data.MapsetId); err != nil {
			return APIErrorServerError("Failed to rank mapset", err)
		}

		if err := db.ResetPersonalBests(data.QueueMapset.Mapset); err != nil {
			return APIErrorServerError("Failed to reset personal best scores for mapset", err)
		}

		if err := db.AddUserActivity(data.QueueMapset.Mapset.CreatorID, db.UserActivityRankedMapset,
			data.QueueMapset.Mapset.String(), data.MapsetId); err != nil {
			return APIErrorServerError("Failed to add new ranked user activity", err)
		}

		if err := db.UpdateElasticSearchMapset(*data.QueueMapset.Mapset); err != nil {
			return APIErrorServerError("Failed to index ranked mapset in elastic search", err)
		}

		if err := db.NewMapsetRankedNotification(data.QueueMapset.Mapset).Insert(); err != nil {
			return APIErrorServerError("Error inserting ranked mapset notification", err)
		}

		_ = webhooks.SendRankedWebhook(data.QueueMapset.Mapset, existingVotes)
	}

	if err := queueMapset.UpdateVoteCount(queueMapset.VoteCount + 1); err != nil {
		return APIErrorServerError("Error updating vote count for queue mapset", err)
	}

	_ = webhooks.SendQueueWebhook(data.User, queueMapset.Mapset, db.RankingQueueActionVote)
	c.JSON(http.StatusOK, gin.H{"message": "You have successfully added a vote to this mapset."})
	return nil
}

// DenyRankingQueueMapset Adds a deny to a ranking queue mapset
// Endpoint: POST /v2/ranking/queue/:id/deny
func DenyRankingQueueMapset(c *gin.Context) *APIError {
	data, apiErr := validateRankingQueueRequest(c)

	if apiErr != nil {
		return apiErr
	}

	queueMapset := data.QueueMapset

	if queueMapset.Mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("This mapset is already ranked.")
	}

	if queueMapset.Status == db.RankingQueueDenied || queueMapset.Status == db.RankingQueueBlacklisted {
		return APIErrorForbidden("This mapset is already denied or blacklisted from the ranking queue.")
	}

	existingDenies, err := db.GetRankingQueueDenies(data.MapsetId)

	if err != nil {
		return APIErrorServerError("Error retrieving ranking queue denies", err)
	}

	for _, denial := range existingDenies {
		if denial.UserId == data.User.Id {
			return APIErrorForbidden("You have already denied this mapset.")
		}

		if denial.User.IsTrialRankingSupervisor() && data.User.IsTrialRankingSupervisor() {
			return APIErrorForbidden("Two trial ranking supervisors cannot deny for the same mapset.")
		}
	}

	denyAction := &db.MapsetRankingQueueComment{
		UserId:     data.User.Id,
		MapsetId:   data.MapsetId,
		ActionType: db.RankingQueueActionDeny,
		IsActive:   true,
		Comment:    data.Comment,
	}

	if err := denyAction.Insert(); err != nil {
		return APIErrorServerError("Error inserting new ranking queue denial", err)
	}

	if err := db.NewMapsetActionNotification(data.QueueMapset.Mapset, denyAction).Insert(); err != nil {
		return APIErrorServerError("Error inserting deny notification", err)
	}

	if len(existingDenies)+1 == config.Instance.RankingQueue.DenialsRequired {
		if err := queueMapset.UpdateStatus(db.RankingQueueDenied); err != nil {
			return APIErrorServerError("Error updating ranking queue mapset status", err)
		}

		if err := queueMapset.UpdateVoteCount(0); err != nil {
			return APIErrorServerError("Error updating vote count for queue mapset", err)
		}

		if err := db.AddUserActivity(data.QueueMapset.Mapset.CreatorID, db.UserActivityDeniedMapset,
			data.QueueMapset.Mapset.String(), data.MapsetId); err != nil {
			return APIErrorServerError("Failed to add new ranked user activity", err)
		}
	}

	_ = webhooks.SendQueueWebhook(data.User, queueMapset.Mapset, db.RankingQueueActionDeny)
	c.JSON(http.StatusOK, gin.H{"message": "You have successfully added a deny to this mapset."})
	return nil
}

// BlacklistRankingQueueMapset Blacklists a mapset from the ranking queue
// Endpoint: POST /v2/ranking/queue/:id/blacklist
func BlacklistRankingQueueMapset(c *gin.Context) *APIError {
	data, apiErr := validateRankingQueueRequest(c)

	if apiErr != nil {
		return apiErr
	}

	queueMapset := data.QueueMapset

	if queueMapset.Mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("This mapset is already ranked.")
	}

	if queueMapset.Status == db.RankingQueueBlacklisted {
		return APIErrorForbidden("This mapset is already blacklisted.")
	}

	blacklistAction := &db.MapsetRankingQueueComment{
		UserId:     data.User.Id,
		MapsetId:   data.MapsetId,
		ActionType: db.RankingQueueActionBlacklist,
		IsActive:   true,
		Comment:    data.Comment,
	}

	if err := blacklistAction.Insert(); err != nil {
		return APIErrorServerError("Error inserting new ranking queue blacklist action", err)
	}

	if err := db.NewMapsetActionNotification(data.QueueMapset.Mapset, blacklistAction).Insert(); err != nil {
		return APIErrorServerError("Error inserting blacklist notification", err)
	}

	if err := queueMapset.UpdateStatus(db.RankingQueueBlacklisted); err != nil {
		return APIErrorServerError("Error updating ranking queue mapset status", err)
	}

	if err := queueMapset.UpdateVoteCount(0); err != nil {
		return APIErrorServerError("Error updating vote count for queue mapset", err)
	}

	_ = webhooks.SendQueueWebhook(data.User, queueMapset.Mapset, db.RankingQueueActionBlacklist)
	c.JSON(http.StatusOK, gin.H{"message": "You have successfully blacklisted this mapset."})
	return nil
}

// OnHoldRankingQueueMapset On-holds  a mapset from the ranking queue
// Endpoint: POST /v2/ranking/queue/:id/hold
func OnHoldRankingQueueMapset(c *gin.Context) *APIError {
	data, apiErr := validateRankingQueueRequest(c)

	if apiErr != nil {
		return apiErr
	}

	queueMapset := data.QueueMapset

	if queueMapset.Mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("This mapset is already ranked.")
	}

	if queueMapset.Status == db.RankingQueueOnHold {
		return APIErrorForbidden("This mapset is already on hold.")
	}

	onHoldAction := &db.MapsetRankingQueueComment{
		UserId:     data.User.Id,
		MapsetId:   data.MapsetId,
		ActionType: db.RankingQueueActionOnHold,
		IsActive:   true,
		Comment:    data.Comment,
	}

	if err := onHoldAction.Insert(); err != nil {
		return APIErrorServerError("Error inserting new ranking queue on hold action.", err)
	}

	if err := db.NewMapsetActionNotification(data.QueueMapset.Mapset, onHoldAction).Insert(); err != nil {
		return APIErrorServerError("Error inserting on hold notification", err)
	}

	if err := queueMapset.UpdateStatus(db.RankingQueueOnHold); err != nil {
		return APIErrorServerError("Error updating ranking queue mapset status", err)
	}

	if err := queueMapset.UpdateVoteCount(0); err != nil {
		return APIErrorServerError("Error updating vote count for queue mapset", err)
	}

	_ = webhooks.SendQueueWebhook(data.User, queueMapset.Mapset, db.RankingQueueActionOnHold)
	c.JSON(http.StatusOK, gin.H{"message": "You have successfully placed this mapset on hold."})
	return nil
}
