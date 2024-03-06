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

const (
	thirtyDays float64 = 24 * 30
)

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
		// User has never changed their name previously.
		if result.Error == gorm.ErrRecordNotFound {
			return true, time.Time{}, result.Error
		}

		return false, time.Time{}, result.Error
	}

	timeSinceLastChange := time.Now().Sub(time.UnixMilli(change.Timestamp))

	if timeSinceLastChange.Hours() < thirtyDays {
		nextChange := time.UnixMilli(change.Timestamp).Add(time.Hour * time.Duration(thirtyDays))
		return false, nextChange, nil
	}

	return true, time.Time{}, nil
}

// IsUsernameAvailable Returns if a username is available to use.
// - A user must not already be using that name
// - A user must not have used that name in the past 60 days.
func IsUsernameAvailable(userId int, username string) (bool, error) {
	user, err := GetUserByUsername(username)

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	// User already has this name
	if user != nil {
		return false, nil
	}

	var change *UsernameChange

	result := SQL.
		Where("user_id != ? AND previous_username = ?", userId, username).
		Order("id DESC").
		Limit(1).
		First(&change)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return true, nil
		}

		return false, err
	}

	// Check if someone has used this username in the past 60 days.
	return time.Now().Sub(time.UnixMilli(change.Timestamp)).Hours() > thirtyDays*2, nil
}
