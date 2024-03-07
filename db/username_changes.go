package db

import (
	"gorm.io/gorm"
	"regexp"
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
	thirtyDays    float64 = 24 * 30
	usernameRegex string  = "^[a-zA-Z0-9](?:[a-zA-Z0-9 ]{0,13}[a-zA-Z0-9])?$"
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
	matched, err := regexp.MatchString(usernameRegex, username)

	if err != nil {
		return false, err
	}

	if !matched {
		return false, nil
	}

	user, err := GetUserByUsername(username)

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	// User already has this name
	if user != nil {
		// A user that isn't them has this name already
		if user.Id != userId {
			return false, nil
		}

		// Disallow user to change it to the same exact username (allows casing change)
		if username == user.Username {
			return false, nil
		}
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

// ChangeUserUsername Changes a user's username
func ChangeUserUsername(userId int, currentName string, username string) (changed bool, unchangedReason string, err error) {
	matched, err := regexp.MatchString(usernameRegex, username)

	if err != nil {
		return false, "", err
	}

	if !matched {
		return false, "Invalid username", nil
	}

	eligible, _, err := CanUserChangeUsername(userId)

	if err != nil {
		return false, "", err
	}

	if !eligible {
		return false, "You must wait at least 30 days before changing your username.", nil
	}

	available, err := IsUsernameAvailable(userId, username)

	if err != nil {
		return false, "", err
	}

	if !available {
		return false, "This username is unavailable.", err
	}

	err = UpdateUserUsername(userId, username)

	if err != nil {
		return false, "", err
	}

	usernameChange := UsernameChange{
		UserId:           userId,
		PreviousUsername: currentName,
		Timestamp:        time.Now().UnixMilli(),
	}

	if err := SQL.Create(&usernameChange).Error; err != nil {
		return false, "", err
	}

	return true, "", nil
}
