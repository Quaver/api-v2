package db

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type ClanActivity struct {
	Id            int              `gorm:"column:id; PRIMARY_KEY" json:"id"`
	ClanId        int              `gorm:"column:clan_id" json:"clan_id"`
	Type          ClanActivityType `gorm:"column:type" json:"type"`
	UserId        int              `gorm:"column:user_id" json:"user_id"`
	MapId         int              `gorm:"column:map_id" json:"map_id"`
	Message       string           `gorm:"column:message" json:"message"`
	Timestamp     int64            `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time        `gorm:"-:all" json:"timestamp"`
}

type ClanActivityType int8

const (
	ClanActivityNone ClanActivityType = iota
	ClanActivityCreated
	ClanActivityUserJoined
	ClanActivityUserLeft
	ClanActivityUserKicked
	ClanActivityOwnershipTransferred
)

func (a *ClanActivity) BeforeCreate(*gorm.DB) (err error) {
	t := time.Now()
	a.TimestampJSON = t
	return nil
}

func (a *ClanActivity) AfterFind(*gorm.DB) (err error) {
	a.TimestampJSON = time.UnixMilli(a.Timestamp)
	return nil
}

// NewClanActivity Creates a new clan activity
func NewClanActivity(clanId int, activityType ClanActivityType, userId int) *ClanActivity {
	return &ClanActivity{ClanId: clanId, Type: activityType, UserId: userId, Timestamp: time.Now().UnixMilli()}
}

// InsertClanActivity Inserts a clan activity into the database
func (a *ClanActivity) InsertClanActivity() error {
	if a.ClanId == 0 {
		return errors.New("cannot insert clan a with no clan id set")
	}

	if a.Type == ClanActivityNone {
		return errors.New("you must set a clan a type before inserting into the database")
	}

	if err := SQL.Create(&a).Error; err != nil {
		return err
	}

	return nil
}

func (*ClanActivity) TableName() string {
	return "clan_activity"
}
