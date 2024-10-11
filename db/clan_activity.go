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
	User          *User            `gorm:"foreignKey:UserId" json:"user"`
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

func (*ClanActivity) TableName() string {
	return "clan_activity"
}

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

// Insert Inserts a clan activity into the database
func (a *ClanActivity) Insert() error {
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

// GetClanActivity Retrieves clan activity from the database
func GetClanActivity(clanId int, limit int, page int) ([]*ClanActivity, error) {
	var activities = make([]*ClanActivity, 0)

	result := SQL.
		Joins("User").
		Where("clan_activity.clan_id = ?", clanId).
		Order("id DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&activities)

	if result.Error != nil {
		return nil, result.Error
	}

	return activities, nil
}
