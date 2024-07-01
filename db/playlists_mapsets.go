package db

type PlaylistMapset struct {
	Id         int            `gorm:"column:id; PRIMARY_KEY" json:"playlist_mapset_id"`
	PlaylistId int            `gorm:"column:playlist_id" json:"-"`
	MapsetId   int            `gorm:"column:mapset_id" json:"-"`
	Mapset     *Mapset        `gorm:"foreignKey:MapsetId" json:"mapset"`
	Maps       []*PlaylistMap `gorm:"foreignKey:PlaylistsMapsetId" json:"maps"`
}

func (*PlaylistMapset) TableName() string {
	return "playlists_mapsets"
}

// GetPlaylistMapsetByIds Gets a playlist mapset by its playlist id and mapset id
func GetPlaylistMapsetByIds(playlistId int, mapsetId int) (*PlaylistMapset, error) {
	var playlistMapset *PlaylistMapset

	result := SQL.
		Where("playlist_id = ? AND mapset_id = ?", playlistId, mapsetId).
		First(&playlistMapset)

	if result.Error != nil {
		return nil, result.Error
	}

	return playlistMapset, nil
}

// Insert Inserts a playlist mapset into the db
func (pm *PlaylistMapset) Insert() error {
	if err := SQL.Create(&pm).Error; err != nil {
		return err
	}

	return nil
}

// DeletePlaylistMapset Deletes a playlist mapset row
func DeletePlaylistMapset(playlistId int, mapsetId int) error {
	result := SQL.Delete(&PlaylistMapset{}, "playlist_id = ? AND mapset_id = ?", playlistId, mapsetId)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetPlaylistMapsetsByMapsetId Gets all the playlist mapsets that have a particular mapset id
func GetPlaylistMapsetsByMapsetId(mapsetId int) ([]*PlaylistMapset, error) {
	var playlistMapsets = make([]*PlaylistMapset, 0)

	result := SQL.
		Preload("Maps").
		Where("mapset_id = ?", mapsetId).
		Find(&playlistMapsets)

	if result.Error != nil {
		return nil, result.Error
	}

	return playlistMapsets, nil
}
