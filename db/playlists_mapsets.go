package db

type PlaylistMapset struct {
	Id         int `gorm:"column:id; PRIMARY_KEY" json:"id"`
	PlaylistId int `gorm:"column:playlist_id" json:"playlist_id"`
	MapsetId   int `gorm:"column:mapset_id" json:"mapset_id"`
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

func (pm *PlaylistMapset) Insert() error {
	if err := SQL.Create(&pm).Error; err != nil {
		return err
	}

	return nil
}
