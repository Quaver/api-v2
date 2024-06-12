package handlers

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// GetRankingQueueConfig Returns the vote/denial configuration for the ranking queue
func GetRankingQueueConfig(c *gin.Context) *APIError {
	c.JSON(http.StatusOK, gin.H{
		"votes_required":          config.Instance.RankingQueue.VotesRequired,
		"denials_required":        config.Instance.RankingQueue.DenialsRequired,
		"mapset_uploads_required": config.Instance.RankingQueue.MapsetUploadsRequired,
		"resubmission_days":       config.Instance.RankingQueue.ResubmissionDays,
	})

	return nil
}

// GetRankingQueue Retrieves the ranking queue for a given mode
// Endpoint: /v2/ranking/queue/mode/:mode
func GetRankingQueue(c *gin.Context) *APIError {
	mode, err := strconv.Atoi(c.Param("mode"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mode")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	rankingQueue, err := db.GetRankingQueue(enums.GameMode(mode), 20, page)

	if err != nil {
		return APIErrorServerError("Error retrieving ranking queue", err)
	}

	c.JSON(http.StatusOK, gin.H{"ranking_queue": rankingQueue})
	return nil
}

// SubmitMapsetToRankingQueue Submits a mapsets to the ranking queue
// Endpoint: POST /v2/ranking/queue/:id/submit
func SubmitMapsetToRankingQueue(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid mapset id")
	}

	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	mapset, err := db.GetMapsetById(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving mapset from db", err)
	}

	if mapset == nil {
		return APIErrorNotFound("Mapset")
	}

	if mapset.CreatorID != user.Id {
		return APIErrorForbidden("You do not own this mapset and cannot submit it to the queue.")
	}

	if len(mapset.Maps) > 0 && mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
		return APIErrorForbidden("This mapset is already ranked.")
	}

	existingQueueMapset, err := db.GetRankingQueueMapset(mapset.Id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving existing mapset in the ranking queue", err)
	}

	uploadedMapsets, err := db.GetUserMapsets(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving user uploaded mapsets from db", err)
	}

	uploadsRequired := config.Instance.RankingQueue.MapsetUploadsRequired

	if len(uploadedMapsets) < uploadsRequired {
		return APIErrorForbidden(fmt.Sprintf("You must have at least %v uploaded mapsets to submit to the queue", uploadsRequired))
	}

	isEligible, err := isEligibleToSubmitToRankingQueue(user.Id, uploadedMapsets)

	if err != nil {
		return APIErrorServerError("Error checking user queue submission eligibility", err)
	}

	if !isEligible {
		return APIErrorForbidden("You cannot submit any more mapsets to the queue at this time.")
	}

	if existingQueueMapset == nil {
		return addMapsetToRankingQueue(c, mapset)
	} else {
		return resubmitMapsetToRankingQueue(c, existingQueueMapset)
	}
}

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
		Comment string `form:"comment" json:"comment"`
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
		MapsetId:   data.MapsetId,
		ActionType: db.RankingQueueActionVote,
		IsActive:   true,
		Comment:    data.Comment,
	}

	if err := newVoteAction.Insert(); err != nil {
		return APIErrorServerError("Error inserting new ranking queue vote", err)
	}

	if len(existingVotes)+1 == config.Instance.RankingQueue.VotesRequired {
		queueMapset.Status = db.RankingQueueRanked

		if err := db.RankMapset(data.MapsetId); err != nil {
			return APIErrorServerError("Failed to rank mapset", err)
		}
	}

	queueMapset.Votes++
	queueMapset.DateLastUpdated = time.Now().UnixMilli()

	if result := db.SQL.Save(queueMapset); result.Error != nil {
		return APIErrorServerError("Error updating ranking queue mapset in database", result.Error)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully added a vote to this mapset."})
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

	queueMapset.Status = db.RankingQueueBlacklisted
	queueMapset.DateLastUpdated = time.Now().UnixMilli()

	if result := db.SQL.Save(queueMapset); result.Error != nil {
		return APIErrorServerError("Error updating ranking queue mapset in database", result.Error)
	}

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

	queueMapset.Status = db.RankingQueueOnHold
	queueMapset.DateLastUpdated = time.Now().UnixMilli()

	if result := db.SQL.Save(data.QueueMapset); result.Error != nil {
		return APIErrorServerError("Error updating ranking queue mapset in database", result.Error)
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully placed this mapset on hold."})
	return nil
}

// Adds a new mapset to the ranking queue
// TODO: Discord Webhook
func addMapsetToRankingQueue(c *gin.Context, mapset *db.Mapset) *APIError {
	rankingQueueMapset := &db.RankingQueueMapset{
		MapsetId:        mapset.Id,
		Timestamp:       time.Now().UnixMilli(),
		DateLastUpdated: time.Now().UnixMilli(),
		Status:          db.RankingQueuePending,
	}

	if err := rankingQueueMapset.Insert(); err != nil {
		return APIErrorServerError("Error inserting mapset into ranking queue", err)
	}

	comment := &db.MapsetRankingQueueComment{
		UserId:     mapset.CreatorID,
		MapsetId:   mapset.Id,
		ActionType: db.RankingQueueActionComment,
		Comment:    "I have just submitted my mapset to the ranking queue!",
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting ranking queue comment", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your mapset was successfully submitted to the queue"})
	return nil
}

// Resubmits a mapset to the ranking queue
// TODO: Discord Webhook
func resubmitMapsetToRankingQueue(c *gin.Context, mapset *db.RankingQueueMapset) *APIError {
	switch mapset.Status {
	case db.RankingQueuePending, db.RankingQueueOnHold, db.RankingQueueResolved:
		return APIErrorForbidden("Your mapset is already pending for rank.")
	case db.RankingQueueBlacklisted:
		return APIErrorForbidden("Your mapset has been blacklisted from the ranking queue.")
	case db.RankingQueueRanked:
		return APIErrorForbidden("Your mapset is already ranked.")
	}

	// Only if mapset is denied will it get to this point.

	resubmitDays := config.Instance.RankingQueue.ResubmissionDays
	timeSinceDenied := time.Now().Sub(time.UnixMilli(mapset.DateLastUpdated))

	if timeSinceDenied.Hours() < 24*float64(resubmitDays) {
		return APIErrorForbidden(fmt.Sprintf("You can only resubmit your mapset for rank every %v days.", resubmitDays))
	}

	mapset.Votes = 0
	mapset.Status = db.RankingQueuePending
	mapset.DateLastUpdated = time.Now().UnixMilli()

	if result := db.SQL.Save(mapset); result.Error != nil {
		return APIErrorServerError("Error updating ranking queue mapset in the database", result.Error)
	}

	// Deactivate previous actions since they no longer count
	if err := db.DeactivateRankingQueueActions(mapset.MapsetId); err != nil {
		return APIErrorServerError("Error deactivating previous ranking queue actions", err)
	}

	comment := &db.MapsetRankingQueueComment{
		UserId:     mapset.Mapset.CreatorID,
		MapsetId:   mapset.Id,
		ActionType: db.RankingQueueActionComment,
		Comment:    "I have just resubmitted my mapset to the ranking queue!",
	}

	if err := comment.Insert(); err != nil {
		return APIErrorServerError("Error inserting ranking queue comment", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your mapset has been resubmitted for rank"})
	return nil
}

// Returns if the user is eligible to submit to the ranking queue
func isEligibleToSubmitToRankingQueue(userId int, uploadedMapsets []*db.Mapset) (bool, error) {
	numRanked := 0

	for _, mapset := range uploadedMapsets {
		if len(mapset.Maps) > 0 && mapset.Maps[0].RankedStatus == enums.RankedStatusRanked {
			numRanked++
		}
	}

	maxSubmissions := 1

	if numRanked >= 5 {
		maxSubmissions = 3
	} else if numRanked >= 2 {
		maxSubmissions = 2
	}

	pendingMapsets, err := db.GetUserMapsetsInRankingQueue(userId)

	if err != nil {
		return false, err
	}

	return len(pendingMapsets) < maxSubmissions, nil
}
