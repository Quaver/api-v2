package db

import (
	"fmt"
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
	Maps                []*MapQua `gorm:"foreignKey:MapsetId" json:"maps,omitempty"`
	User                *User     `gorm:"foreignKey:CreatorID; references:Id" json:"user,omitempty"`
}

func (m *Mapset) TableName() string {
	return "mapsets"
}

func (m *Mapset) String() string {
	return fmt.Sprintf("%v - %v", m.Artist, m.Title)
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

func (m *Mapset) Insert() error {
	m.IsVisible = true
	m.DateSubmittedJSON = time.Now()
	m.DateLastUpdatedJSON = time.Now()

	if err := SQL.Create(&m).Error; err != nil {
		return err
	}

	return nil
}

// GetMapsetById Retrieves a mapset by its id
func GetMapsetById(id int) (*Mapset, error) {
	var mapset *Mapset

	result := SQL.
		Joins("User").
		Preload("Maps").
		Where("mapsets.id = ? AND mapsets.visible = 1", id).
		First(&mapset)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := mapset.User.AfterFind(SQL); err != nil {
		return nil, err
	}

	return mapset, nil
}

// GetUserMapsets Retrieves a user's uploaded mapsets
func GetUserMapsets(userId int) ([]*Mapset, error) {
	var mapsets = make([]*Mapset, 0)

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

// GetUserMonthlyUploadMapsets Retrieves a user's mapsets that they've uploaded in the past month
func GetUserMonthlyUploadMapsets(userId int) ([]*Mapset, error) {
	var mapsets = make([]*Mapset, 0)

	thirtyDays := int64(1000 * 60 * 60 * 24 * 30)
	monthAgo := time.Now().UnixMilli() - thirtyDays

	result := SQL.
		Preload("Maps").
		Where("mapsets.creator_id = ? AND "+
			"mapsets.date_submitted > ?", userId, monthAgo).
		Find(&mapsets)

	if result.Error != nil {
		return nil, result.Error
	}

	return mapsets, nil
}

// GetAllMapsets Retrieves all the mapsets in the database
func GetAllMapsets() ([]*Mapset, error) {
	var mapsets = make([]*Mapset, 0)

	result := SQL.
		Preload("Maps").
		Where("mapsets.visible = 1").
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
	var ids = make([]int, 0)

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

	var offsets = make([]*onlineOffset, 0)

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

// ResetPersonalBests Resets the personal best scores of all maps in a set.
// Usually used when ranking a mapset
func ResetPersonalBests(mapset *Mapset) error {
	for _, songMap := range mapset.Maps {
		result := SQL.Model(&Score{}).
			Where("map_md5 = ?", songMap.MD5).
			Update("personal_best", 0)

		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

// DeleteMapset Deletes (hides) a given mapset
func DeleteMapset(id int) error {
	result := SQL.Model(&Mapset{}).Where("id = ?", id).Update("visible", 0)

	if result.Error != nil {
		return result.Error
	}

	if err := DeleteElasticSearchMapset(id); err != nil {
		return err
	}

	return nil
}

// UpdateMapsetPackageMD5 Updates the package md5 of a mapset
func UpdateMapsetPackageMD5(id int, md5 string) error {
	result := SQL.Model(&Mapset{}).
		Where("id = ?", id).
		Update("package_md5", md5)

	return result.Error
}

// UpdateMetadata Updates the metadata of a given mapset (username, artist, title, etc)
func (m *Mapset) UpdateMetadata() error {
	result := SQL.Model(&Mapset{}).
		Where("id = ?", m.Id).
		Updates(map[string]interface{}{
			"creator_username":  m.CreatorUsername,
			"artist":            m.Artist,
			"title":             m.Title,
			"source":            m.Source,
			"tags":              m.Tags,
			"date_last_updated": time.Now().UnixMilli(),
		})

	return result.Error
}
