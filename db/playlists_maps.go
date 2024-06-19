package db

type PlaylistMap struct {
	Id                int     `gorm:"column:id; PRIMARY_KEY" json:"id"`
	PlaylistId        int     `gorm:"column:playlist_id" json:"playlist_id"`
	MapId             int     `gorm:"column:map_id" json:"map_id"`
	PlaylistsMapsetId int     `gorm:"column:playlists_mapsets_id" json:"mapset_id"`
	Map               *MapQua `gorm:"foreignKey:MapId" json:"-"`
}

func (*PlaylistMap) TableName() string {
	return "playlists_maps"
}
