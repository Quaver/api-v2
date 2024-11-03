package db

import (
	"errors"
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/stringutil"
	"gorm.io/gorm"
	"regexp"
	"time"
)

type Clan struct {
	Id              int          `gorm:"column:id; PRIMARY_KEY" json:"id"`
	OwnerId         int          `gorm:"column:owner_id" json:"owner_id"`
	Name            string       `gorm:"column:name" json:"name"`
	Tag             string       `gorm:"column:tag" json:"tag"`
	CreatedAt       int64        `gorm:"column:created_at" json:"-"`
	AboutMe         *string      `gorm:"column:about_me" json:"about_me"`
	FavoriteMode    uint8        `gorm:"column:favorite_mode" json:"favorite_mode"`
	AccentColor     *string      `gorm:"column:accent_color" json:"accent_color"`
	LastUpdated     int64        `gorm:"column:last_updated" json:"-"`
	IsCustomizable  bool         `gorm:"column:customizable" json:"customizable"`
	CreatedAtJSON   time.Time    `gorm:"-:all" json:"created_at"`
	LastUpdatedJSON time.Time    `gorm:"-:all" json:"last_updated"`
	Stats           []*ClanStats `gorm:"foreignKey:ClanId" json:"stats"`
}

func (clan *Clan) BeforeCreate(*gorm.DB) (err error) {
	clan.CreatedAtJSON = time.Now()
	clan.LastUpdatedJSON = time.Now()
	return nil
}

func (clan *Clan) AfterFind(*gorm.DB) (err error) {
	clan.CreatedAtJSON = time.UnixMilli(clan.CreatedAt)
	clan.LastUpdatedJSON = time.UnixMilli(clan.LastUpdated)
	return nil
}

func (clan *Clan) AvatarURL() string {
	return fmt.Sprintf("https://cdn.quavergame.com/clan-avatars/%v.jpg", clan.Id)
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
	clan.LastUpdated = time.Now().UnixMilli()

	err := SQL.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&clan).Error; err != nil {
			return err
		}

		for i := 1; i <= 2; i++ {
			stat := &ClanStats{ClanId: clan.Id, Mode: enums.GameMode(i)}

			if err := tx.Create(stat).Error; err != nil {
				return err
			}

			clan.Stats = append(clan.Stats, stat)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetClanById Gets a clan from the database by its id
func GetClanById(id int) (*Clan, error) {
	var clan *Clan

	result := SQL.
		Preload("Stats").
		Where("clans.id = ?", id).First(&clan)

	if result.Error != nil {
		return nil, result.Error
	}

	return clan, nil
}

// GetClanByName Gets a clan from the database by its name
func GetClanByName(name string) (*Clan, error) {
	var clan *Clan

	result := SQL.
		Preload("Stats").
		Where("clans.name = ?", name).First(&clan)

	if result.Error != nil {
		return nil, result.Error
	}

	return clan, nil
}

// GetClanByTag Gets a clan from the db by its tag (case-sensitive)
func GetClanByTag(tag string) (*Clan, error) {
	var clan *Clan

	result := SQL.
		Preload("Stats").
		Where("BINARY clans.tag = ?", tag).First(&clan)

	if result.Error != nil {
		return nil, result.Error
	}

	return clan, nil
}

// GetClansCount Gets the total amount of clans
func GetClansCount() (int, error) {
	var count int

	result := SQL.Raw("SELECT COUNT(*) as count FROM clans").
		Scan(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetClanMemberCount Returns the amount of users that are in a given clan
func GetClanMemberCount(clanId int) (int, error) {
	var count int

	result := SQL.Raw("SELECT COUNT(*) as count FROM users WHERE clan_id = ?", clanId).
		Scan(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetClanTagAndAccentColor Retrieves the name and tag for a clan
func GetClanTagAndAccentColor(clanId int) (string, *string, error) {
	var data Clan

	result := SQL.
		Raw("SELECT tag, accent_color FROM clans WHERE id = ?", clanId).
		Scan(&data)

	if result.Error != nil {
		return "", nil, result.Error
	}

	return data.Tag, data.AccentColor, nil
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

		if err := tx.Model(&Score{}).Where("clan_id = ?", id).Update("clan_id", nil).Error; err != nil {
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

func DoesClanExistByTag(tag string) (bool, *Clan, error) {
	clan, err := GetClanByTag(tag)

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, nil, err
	}

	return clan != nil, clan, nil
}

// IsValidClanName Checks a string to see if it is a valid clan name
func IsValidClanName(name string) bool {
	result, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9 ]{2,29}$", name)
	return result
}

// IsValidClanTag Checks a string to see if it is a valid clan tag
func IsValidClanTag(tag string) bool {
	if stringutil.IsClanTagCensored(tag) {
		return false
	}

	result, _ := regexp.MatchString("^[a-zA-Z0-9]{1,4}$", tag)
	return result
}

// UpdateOwner Updates the owner of a clan
func (clan *Clan) UpdateOwner(ownerId int) error {
	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("owner_id", ownerId)

	return result.Error
}

// UpdateName Updates the name of a clan
func (clan *Clan) UpdateName(name string) error {
	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("name", name)

	return result.Error
}

// UpdateTag Updates the tag of a clan
func (clan *Clan) UpdateTag(tag string) error {
	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("tag", tag)

	return result.Error
}

// UpdateFavoriteMode Updates the favorite mode of a clan
func (clan *Clan) UpdateFavoriteMode(mode enums.GameMode) error {
	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("favorite_mode", mode)

	return result.Error
}

// UpdateAboutMe Updates a clan's about me
func (clan *Clan) UpdateAboutMe(aboutMe string) error {
	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("about_me", aboutMe)

	return result.Error
}

// UpdateCustomizable Updates a clan's customizability status
func (clan *Clan) UpdateCustomizable(enabled bool) error {
	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("customizable", enabled)

	return result.Error
}

// UpdateAccentColor Updates the accent color for a clan
func (clan *Clan) UpdateAccentColor(hex string) error {
	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("accent_color", hex)

	return result.Error
}

func (clan *Clan) UpdateLastUpdated() error {
	clan.LastUpdated = time.Now().UnixMilli()
	clan.LastUpdatedJSON = time.Now()

	result := SQL.Model(&Clan{}).
		Where("id = ?", clan.Id).
		Update("last_updated", clan.LastUpdated)

	return result.Error
}
