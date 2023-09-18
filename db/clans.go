package db

import (
	"errors"
	"regexp"
	"time"
)

type Clan struct {
	Id                 uint    `gorm:"column:id; PRIMARY_KEY"`
	OwnerId            int     `gorm:"column:owner_id"`
	Name               string  `gorm:"column:name"`
	Tag                string  `gorm:"column:tag"`
	CreatedAt          int64   `gorm:"column:created_at"`
	AboutMe            *string `gorm:"column:about_me"`
	FavoriteMode       uint8   `gorm:"column:favorite_mode"`
	LastNameChangeTime int64   `gorm:"column:last_name_change_time"`
}

// GetClanByName Gets a clan from the database by its name
func GetClanByName(name string) (*Clan, error) {
	var clan *Clan

	result := SQL.Where("name = ?", name).First(&clan)

	if result.Error != nil {
		return nil, result.Error
	}

	return clan, nil
}

// Insert Inserts the clan into the database
func (clan *Clan) Insert() error {
	if clan.Id != 0 {
		return errors.New("cannot insert clan that already exists in the database")
	}

	if clan.OwnerId == 0 {
		return errors.New("cannot insert clan without a clan owner")
	}

	clan.FavoriteMode = 1
	clan.CreatedAt = time.Now().UnixMilli()
	clan.LastNameChangeTime = time.Now().UnixMilli()

	result := SQL.Create(&clan)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// IsValidClanName Checks a string to see if it is a valid clan name
func IsValidClanName(name string) bool {
	result, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9 ]{2,29}$", name)
	return result
}

// IsValidClanTag Checks a string to see if it is a valid clan tag
func IsValidClanTag(tag string) bool {
	result, _ := regexp.MatchString("^[a-zA-Z0-9]{1,4}$", tag)
	return result
}
