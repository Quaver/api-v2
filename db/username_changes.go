package db

import (
	"gorm.io/gorm"
	"time"
)

type UsernameChange struct {
	Id               int    `gorm:"column:id; PRIMARY_KEY"`
	UserId           int    `gorm:"column:user_Id"`
	PreviousUsername string `gorm:"previous_username"`
	Timestamp        int64  `gorm:"timestamp"`
}

func (*UsernameChange) TableName() string {
	return "username_changes"
}

// CanUserChangeUsername Checks and returns if the user is allowed to change their username.
// When a user changes their username, they are required to wait at least 30 days before they can
// change it again. Returns if the user can change, the time left to change, and an error.
func CanUserChangeUsername(userId int) (bool, time.Time, error) {
	var change *UsernameChange

	result := SQL.
		Where("user_id = ?", userId).
		Order("id DESC").
		Limit(1).
		First(&change)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return true, time.Time{}, result.Error
		}

		return false, time.Time{}, result.Error
	}

	timeSinceLastChange := time.Now().Sub(time.UnixMilli(change.Timestamp))
	const thirtyDays float64 = 24 * 30

	if timeSinceLastChange.Hours() < thirtyDays {
		nextChange := time.UnixMilli(change.Timestamp).Add(time.Hour * time.Duration(thirtyDays))
		return false, nextChange, nil
	}

	return true, time.Time{}, nil
}
