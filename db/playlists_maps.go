package db

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
