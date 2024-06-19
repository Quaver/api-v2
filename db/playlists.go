package db

import (
	"gorm.io/gorm"
	"time"
)

type Playlist struct {
	Id                  int            `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId              int            `gorm:"column:user_id" json:"user_id"`
	Name                string         `gorm:"column:name" json:"name"`
	Description         string         `gorm:"column:description" json:"description"`
	LikeCount           int            `gorm:"column:like_count" json:"-"`
	MapCount            int            `gorm:"column:map_count" json:"map_count"`
	Timestamp           int64          `gorm:"column:timestamp" json:"-"`
	TimestampJSON       time.Time      `gorm:"-:all" json:"timestamp"`
	TimeLastUpdated     int64          `gorm:"column:time_last_updated" json:"-"`
	TimeLastUpdatedJSON time.Time      `gorm:"-:all" json:"time_last_updated"`
	Visible             bool           `gorm:"column:visible" json:"-"`
	Maps                []*PlaylistMap `gorm:"foreignKey:PlaylistId" json:"-"`
}

func (*Playlist) TableName() string {
	return "playlists"
}

func (p *Playlist) AfterFind(*gorm.DB) (err error) {
	p.TimestampJSON = time.UnixMilli(p.Timestamp)
	p.TimeLastUpdatedJSON = time.UnixMilli(p.TimeLastUpdated)
	return nil
}

// GetAllPlaylists Returns all the playlists in the db
func GetAllPlaylists() ([]*Playlist, error) {
	var playlists []*Playlist

	result := SQL.
		Preload("Maps").
		Preload("Maps.Map").
		Find(&playlists)

	if result.Error != nil {
		return nil, result.Error
	}

	return playlists, nil
}
