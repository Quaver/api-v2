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
	if err := SQL.Create(&c).Error; err != nil {
		return err
	}

	return nil
}

// GetRankingQueueComments Retrieves the ranking queue comments for a given mapset
func GetRankingQueueComments(mapsetId int) ([]*MapsetRankingQueueComment, error) {
	var comments []*MapsetRankingQueueComment

	result := SQL.
		Joins("User").
		Where("mapset_id = ?", mapsetId).
		Order("Id ASC").
		Find(&comments)

	if result.Error != nil {
		return nil, result.Error
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
