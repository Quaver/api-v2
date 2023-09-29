package db

import (
	"gorm.io/gorm"
	"time"
)

type ClanInvite struct {
	Id            int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	ClanId        int       `gorm:"column:clan_id" json:"clan_id"`
	UserId        int       `gorm:"column:user_id" json:"user_id"`
	CreatedAt     int64     `gorm:"column:created_at" json:"-"`
	CreatedAtJSON time.Time `gorm:"-:all" json:"created_at"`
}

func (*ClanInvite) TableName() string {
	return "clan_invites"
}

func (invite *ClanInvite) BeforeCreate(*gorm.DB) (err error) {
	invite.CreatedAtJSON = time.Now()
	return nil
}

func (invite *ClanInvite) AfterFind(*gorm.DB) (err error) {
	invite.CreatedAtJSON = time.UnixMilli(invite.CreatedAt)
	return nil
}

// GetPendingClanInvite Retrieves a clan invite for a given user
func GetPendingClanInvite(clanId int, userId int) (*ClanInvite, error) {
	var invite *ClanInvite

	result := SQL.Where("clan_id = ? AND user_id = ?", clanId, userId).First(&invite)

	if result.Error != nil {
		return nil, result.Error
	}

	return invite, nil
}

// InviteUserToClan Creates a clan invite for a user
func InviteUserToClan(clanId int, userId int) (*ClanInvite, error) {
	invite := &ClanInvite{ClanId: clanId, UserId: userId, CreatedAt: time.Now().UnixMilli()}

	if err := SQL.Create(&invite).Error; err != nil {
		return nil, err
	}

	return invite, nil
}

// GetClanInviteById GetClanInvite Retrieves a clan invite at a specific id
func GetClanInviteById(id int) (*ClanInvite, error) {
	var invite *ClanInvite

	result := SQL.Where("id = ?", id).First(&invite)

	if result.Error != nil {
		return nil, result.Error
	}

	return invite, nil
}

// GetUserClanInvites Retrieves a list of pending clan invites for the user
func GetUserClanInvites(userId int) ([]*ClanInvite, error) {
	var invites []*ClanInvite

	result := SQL.Where("user_id = ?", userId).Find(&invites)

	if result.Error != nil {
		return nil, result.Error
	}

	return invites, nil
}

// DeleteClanInviteById Deletes an individual clan invite
func DeleteClanInviteById(id int) error {
	if err := SQL.Delete(&ClanInvite{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}
