package db

import (
	"gorm.io/gorm"
)

type PlaylistMap struct {
	Id                int     `gorm:"column:id; PRIMARY_KEY" json:"playlist_map_id"`
	PlaylistId        int     `gorm:"column:playlist_id" json:"-"`
	MapId             int     `gorm:"column:map_id" json:"-"`
	PlaylistsMapsetId int     `gorm:"column:playlists_mapsets_id" json:"-"`
	Map               *MapQua `gorm:"foreignKey:MapId" json:"map"`
}

func (*PlaylistMap) TableName() string {
	return "playlists_maps"
}

// DoesPlaylistContainMap Checks if a playlist contains a specific map id
func DoesPlaylistContainMap(playlistId int, mapId int) (bool, error) {
	var playlistMap *PlaylistMap

	result := SQL.
		Where("playlist_id = ? AND map_id = ?", playlistId, mapId).
		First(&playlistMap)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, result.Error
	}

	return true, nil
}

// Insert Inserts a playlist map into the db
func (pm *PlaylistMap) Insert() error {
	if err := SQL.Create(&pm).Error; err != nil {
		return err
	}

	return nil
}

// DeletePlaylistMap Deletes a playlist map row
func DeletePlaylistMap(playlistId int, mapId int) error {
	result := SQL.Delete(&PlaylistMap{}, "playlist_id = ? AND map_id = ?", playlistId, mapId)

	if err := result.Error; err != nil {
		return err
	}

	return nil
}
