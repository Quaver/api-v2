package db

import (
	"errors"
	"regexp"
	"time"
)

type Clan struct {
	Id                     int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	OwnerId                int       `gorm:"column:owner_id" json:"owner_id"`
	Name                   string    `gorm:"column:name" json:"name"`
	Tag                    string    `gorm:"column:tag" json:"tag"`
	CreatedAt              int64     `gorm:"column:created_at" json:"-"`
	CreatedAtJSON          time.Time `gorm:"-:all" json:"created_at"`
	AboutMe                *string   `gorm:"column:about_me" json:"about_me"`
	FavoriteMode           uint8     `gorm:"column:favorite_mode" json:"favorite_mode"`
	LastNameChangeTime     int64     `gorm:"column:last_name_change_time" json:"-"`
	LastNameChangeTimeJSON time.Time `gorm:"-:all" json:"last_name_change_time"`
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
	clan.CreatedAtJSON = time.Now()
	clan.LastNameChangeTimeJSON = time.Now()

	result := SQL.Create(&clan)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetClanByName Gets a clan from the database by its name
func GetClanByName(name string) (*Clan, error) {
	var clan *Clan

	result := SQL.Where("name = ?", name).First(&clan)

	if result.Error != nil {
		return nil, result.Error
	}

	clan.CreatedAtJSON = time.UnixMilli(clan.CreatedAt)
	clan.LastNameChangeTimeJSON = time.UnixMilli(clan.LastNameChangeTime)
	return clan, nil
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
