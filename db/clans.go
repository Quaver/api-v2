package db

import (
	"errors"
	"gorm.io/gorm"
	"regexp"
	"time"
)

type Clan struct {
	Id                     int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	OwnerId                int       `gorm:"column:owner_id" json:"owner_id"`
	Name                   string    `gorm:"column:name" json:"name"`
	Tag                    string    `gorm:"column:tag" json:"tag"`
	CreatedAt              int64     `gorm:"column:created_at" json:"-"`
	AboutMe                *string   `gorm:"column:about_me" json:"about_me"`
	FavoriteMode           uint8     `gorm:"column:favorite_mode" json:"favorite_mode"`
	LastNameChangeTime     int64     `gorm:"column:last_name_change_time" json:"-"`
	CreatedAtJSON          time.Time `gorm:"-:all" json:"created_at"`
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

	err := SQL.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&clan).Error; err != nil {
			return err
		}

		for i := 1; i <= 2; i++ {
			if err := tx.Create(&ClanStats{ClanId: clan.Id, Mode: i}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// BeforeCreate Updates clan timestamps before inserting into the db
func (clan *Clan) BeforeCreate(*gorm.DB) (err error) {
	clan.CreatedAtJSON = time.Now()
	clan.LastNameChangeTimeJSON = time.Now()
	return nil
}

// AfterFind Updates clan timestamps after selecting in the db
func (clan *Clan) AfterFind(*gorm.DB) (err error) {
	clan.CreatedAtJSON = time.UnixMilli(clan.CreatedAt)
	clan.LastNameChangeTimeJSON = time.UnixMilli(clan.LastNameChangeTime)
	return nil
}

// GetClanById Gets a clan from the database by its id
func GetClanById(id int) (*Clan, error) {
	var clan *Clan

	result := SQL.Where("id = ?", id).First(&clan)

	if result.Error != nil {
		return nil, result.Error
	}

	return clan, nil
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

// DeleteClan Fully deletes a clan with a given id
func DeleteClan(id int) error {
	err := SQL.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&Clan{}, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&ClanStats{}, "clan_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&ClanActivity{}, "clan_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&ClanInvite{}, "clan_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&ClanScore{}, "clan_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Model(&User{}).Where("clan_id = ?", id).Update("clan_id", nil).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// DoesClanExistByName Returns if a clan exists by its name
func DoesClanExistByName(name string) (bool, error) {
	clan, err := GetClanByName(name)

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	return clan != nil, nil
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
