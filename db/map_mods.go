package db

import (
	"gorm.io/gorm"
	"time"
)

type MapModStatus string
type MapModType string

const (
	ModStatusPending  MapModStatus = "Pending"
	ModStatusAccepted MapModStatus = "Accepted"
	ModStatusDenied   MapModStatus = "Denied"
	ModStatusIgnored  MapModStatus = "Ignored"

	ModTypeIssue      MapModType = "Issue"
	ModTypeSuggestion MapModType = "Suggestion"
)

type MapMod struct {
	Id            int              `gorm:"column:id; PRIMARY_KEY" json:"id"`
	MapId         int              `gorm:"column:map_id" json:"map_id"`
	AuthorId      int              `gorm:"column:author_id" json:"author_id"`
	Timestamp     int64            `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time        `gorm:"-:all" json:"timestamp"`
	MapTimestamp  *string          `gorm:"column:map_timestamp" json:"map_timestamp"`
	Comment       string           `gorm:"column:comment" json:"comment"`
	Status        MapModStatus     `gorm:"column:status" json:"status"`
	Type          MapModType       `gorm:"column:type" json:"type"`
	Author        *User            `gorm:"foreignKey:AuthorId; references:Id" json:"author"`
	Replies       []*MapModComment `gorm:"foreignKey:MapModId" json:"replies"`
}

func (*MapMod) TableName() string {
	return "map_mods"
}

func (mod *MapMod) AfterFind(*gorm.DB) (err error) {
	mod.TimestampJSON = time.UnixMilli(mod.Timestamp)
	return nil
}

// GetMapMods Retrieves map mods for a given map
func GetMapMods(id int) ([]*MapMod, error) {
	var mods = make([]*MapMod, 0)

	result := SQL.
		Joins("Author").
		Preload("Replies").
		Preload("Replies.Author").
		Where("map_id = ?", id).
		Find(&mods)

	if result.Error != nil {
		return nil, result.Error
	}

	return mods, nil
}

// GetModById Gets a mod by its id
func GetModById(id int) (*MapMod, error) {
	var mod *MapMod

	result := SQL.
		Joins("Author").
		Preload("Replies").
		Preload("Replies.Author").
		Where("map_mods.id = ?", id).
		First(&mod)

	if result.Error != nil {
		return nil, result.Error
	}

	return mod, nil
}

// Insert Inserts a new mod into the database
func (mod *MapMod) Insert() error {
	mod.Status = ModStatusPending
	mod.Timestamp = time.Now().UnixMilli()

	if err := SQL.Create(&mod).Error; err != nil {
		return err
	}

	return nil
}
