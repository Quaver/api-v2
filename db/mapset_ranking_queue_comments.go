package db

import (
	"gorm.io/gorm"
	"time"
)

type RankingQueueAction int8

const (
	RankingQueueActionComment RankingQueueAction = iota
	RankingQueueActionDeny
	RankingQueueActionBlacklist
	RankingQueueActionOnHold
	RankingQueueActionVote
)

type MapsetRankingQueueComment struct {
	Id                  int                `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId              int                `gorm:"column:user_id" json:"user_id"`
	MapsetId            int                `gorm:"column:mapset_id" json:"mapset_id"`
	ActionType          RankingQueueAction `gorm:"action_type" json:"action_type"`
	IsActive            bool               `gorm:"is_active" json:"is_active"` // If action counts toward ranking
	Timestamp           int64              `gorm:"column:timestamp" json:"-"`
	TimestampJSON       time.Time          `gorm:"-:all" json:"timestamp"`
	Comment             string             `gorm:"comment" json:"comment"`
	DateLastUpdated     int64              `gorm:"date_last_updated" json:"-"`
	DateLastUpdatedJSON time.Time          `gorm:"-:all" json:"date_last_updated"`
	User                *User              `gorm:"foreignKey:UserId; references:Id" json:"user"`
}

func (*MapsetRankingQueueComment) TableName() string {
	return "mapset_ranking_queue_comments"
}

func (c *MapsetRankingQueueComment) AfterFind(*gorm.DB) (err error) {
	c.TimestampJSON = time.UnixMilli(c.Timestamp)
	c.DateLastUpdatedJSON = time.UnixMilli(c.DateLastUpdated)
	return nil
}

// Insert Inserts a ranking queue comment into the database
func (c *MapsetRankingQueueComment) Insert() error {
	c.Timestamp = time.Now().UnixMilli()
	c.DateLastUpdated = time.Now().UnixMilli()

	if err := SQL.Create(&c).Error; err != nil {
		return err
	}

	return nil
}

// GetRankingQueueComments Retrieves the ranking queue comments for a given mapset
func GetRankingQueueComments(mapsetId int) ([]*MapsetRankingQueueComment, error) {
	var comments = make([]*MapsetRankingQueueComment, 0)

	result := SQL.
		Joins("User").
		Where("mapset_id = ?", mapsetId).
		Order("Id ASC").
		Find(&comments)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, comment := range comments {
		if err := comment.User.AfterFind(SQL); err != nil {
			return nil, err
		}
	}

	return comments, nil
}

// GetRankingQueueComment Retrieves a ranking queue comment at a given id
func GetRankingQueueComment(id int) (*MapsetRankingQueueComment, error) {
	var comment *MapsetRankingQueueComment

	result := SQL.
		Where("id = ?", id).
		First(&comment)

	if result.Error != nil {
		return nil, result.Error
	}

	return comment, nil
}

// DeactivateRankingQueueActions "De-activates" ranking queue actions.
// This means that these specific actions do not currently count towards ranking.
// Example - If a mapset gets denied, all previous actions (votes/denies) no longer count
func DeactivateRankingQueueActions(mapsetId int) error {
	result := SQL.Model(&MapsetRankingQueueComment{}).
		Where("mapset_id = ?", mapsetId).
		Updates(map[string]interface{}{
			"is_active": false,
		})

	return result.Error
}

// GetRankingQueueVotes Retrieves the active votes for a given mapset in the ranking queue
func GetRankingQueueVotes(mapsetId int) ([]*MapsetRankingQueueComment, error) {
	var votes = make([]*MapsetRankingQueueComment, 0)

	result := SQL.
		Joins("User").
		Where("mapset_id = ? AND action_type = ? AND is_active = 1", mapsetId, RankingQueueActionVote).
		Find(&votes)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, vote := range votes {
		if err := vote.User.AfterFind(SQL); err != nil {
			return nil, err
		}
	}

	return votes, nil
}

// GetRankingQueueDenies Retrieves the active denies for a given mapset in the ranking queue
func GetRankingQueueDenies(mapsetId int) ([]*MapsetRankingQueueComment, error) {
	var votes = make([]*MapsetRankingQueueComment, 0)

	result := SQL.
		Joins("User").
		Where("mapset_id = ? AND action_type = ? AND is_active = 1", mapsetId, RankingQueueActionDeny).
		Find(&votes)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, vote := range votes {
		if err := vote.User.AfterFind(SQL); err != nil {
			return nil, err
		}
	}

	return votes, nil
}
