package db

import (
	"gorm.io/gorm"
	"time"
)

type UserActivity struct {
	Id            int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId        int       `gorm:"column:user_id" json:"user_id"`
	Type          int8      `gorm:"column:type" json:"type"`
	Timestamp     int64     `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time `gorm:"-:all" json:"timestamp"`
	Value         string    `gorm:"column:value" json:"value"`
	MapsetId      int       `gorm:"mapset_id" json:"mapset_id"`
}

func (*UserActivity) TableName() string {
	return "activity_feed"
}

func (ua *UserActivity) BeforeCreate(*gorm.DB) (err error) {
	ua.TimestampJSON = time.Now()
	return nil
}

func (ua *UserActivity) AfterFind(*gorm.DB) (err error) {
	ua.TimestampJSON = time.UnixMilli(ua.Timestamp)
	return nil
}

// GetRecentUserActivity Gets the most recent user activity
func GetRecentUserActivity(id int, limit int, page int) ([]*UserActivity, error) {
	var activity []*UserActivity

	result := SQL.
		Where("user_id = ?", id).
		Order("timestamp DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&activity)

	if result.Error != nil {
		return nil, result.Error
	}

	return activity, nil
}