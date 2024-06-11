package db

import (
	"github.com/Quaver/api2/enums"
	"gorm.io/gorm"
	"time"
)

type Mapset struct {
	Id                  int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	PackageMD5          string    `gorm:"column:package_md5" json:"package_md5"`
	CreatorID           int       `gorm:"column:creator_id" json:"creator_id"`
	CreatorUsername     string    `gorm:"column:creator_username" json:"creator_username"`
	Artist              string    `gorm:"column:artist" json:"artist"`
	Title               string    `gorm:"column:title" json:"title"`
	Source              string    `gorm:"column:source" json:"source"`
	Tags                string    `gorm:"column:tags" json:"tags"`
	Description         string    `gorm:"column:description" json:"description"`
	DateSubmitted       int64     `gorm:"column:date_submitted" json:"-"`
	DateSubmittedJSON   time.Time `gorm:"-:all" json:"date_submitted"`
	DateLastUpdated     int64     `gorm:"column:date_last_updated" json:"-"`
	DateLastUpdatedJSON time.Time `gorm:"-:all" json:"date_last_updated"`
	IsVisible           bool      `gorm:"column:visible" json:"is_visible"`
	Maps                []*MapQua `gorm:"foreignKey:MapsetId" json:"maps"`
}

func (m *Mapset) TableName() string {
	return "mapsets"
}

func (m *Mapset) BeforeCreate(*gorm.DB) (err error) {
	m.DateSubmittedJSON = time.Now()
	m.DateLastUpdatedJSON = time.Now()
	return nil
}

func (m *Mapset) AfterFind(*gorm.DB) (err error) {
	m.DateSubmittedJSON = time.UnixMilli(m.DateSubmitted)
	m.DateLastUpdatedJSON = time.UnixMilli(m.DateLastUpdated)
	return nil
}

// GetMapsetById Retrieves a mapset by its id
func GetMapsetById(id int) (*Mapset, error) {
	var mapset *Mapset

	result := SQL.
		Preload("Maps").
		Where("mapsets.id = ? AND mapsets.visible = 1", id).
		First(&mapset)

	if result.Error != nil {
		return nil, result.Error
	}

	return mapset, nil
}

// GetUserMapsets Retrieves a user's uploaded mapsets
func GetUserMapsets(userId int) ([]*Mapset, error) {
	var mapsets []*Mapset

	result := SQL.
		Preload("Maps").
		Where("mapsets.creator_id = ? AND "+
			"mapsets.visible = 1", userId).
		Order("mapsets.date_last_updated DESC").
		Find(&mapsets)

	if result.Error != nil {
		return nil, result.Error
	}

	return mapsets, nil
}

// UpdateMapsetDescription Updates a given mapset's description
func UpdateMapsetDescription(id int, description string) error {
	result := SQL.Model(&Mapset{}).Where("id = ?", id).Update("description", description)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetRankedMapsetIds Retrieves a list of ranked mapset ids
func GetRankedMapsetIds() ([]int, error) {
	var ids []int

	result := SQL.Raw("SELECT DISTINCT ms.id FROM mapsets AS ms " +
		"INNER JOIN maps AS m ON m.mapset_id = ms.id " +
		"WHERE m.ranked_status = 2 AND ms.visible = 1").
		Scan(&ids)

	if result.Error != nil {
		return nil, result.Error
	}

	return ids, nil
}

// GetMapsetOnlineOffsets Retrieves a list of online offsets
func GetMapsetOnlineOffsets() (interface{}, error) {
	type onlineOffset struct {
		Id           int `gorm:"column:id" json:"id"`
		OnlineOffset int `gorm:"column:online_offset" json:"offset"`
	}

	var offsets []*onlineOffset

	result := SQL.Raw("SELECT m.id, m.online_offset FROM maps m " +
		"INNER JOIN mapsets ms ON m.mapset_id = ms.id " +
		"WHERE m.ranked_status = 2 AND ms.visible = 1 AND m.online_offset != 0").
		Scan(&offsets)

	if result.Error != nil {
		return nil, result.Error
	}

	return offsets, nil
}

// RankMapset Ranks all maps in a mapset
func RankMapset(id int) error {
	result := SQL.Model(&MapQua{}).
		Where("mapset_id = ?", id).
		Update("ranked_status", enums.RankedStatusRanked)

	if result.Error != nil {
		return result.Error
	}

	result = SQL.Model(&Mapset{}).
		Where("id = ?", id).
		Update("date_last_updated", time.Now().UnixMilli())

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteMapset Deletes (hides) a given mapset
func DeleteMapset(id int) error {
	result := SQL.Model(&Mapset{}).Where("id = ?", id).Update("visible", 0)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
