package db

type MusicArtistAlbum struct {
	Id        int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	ArtistId  int    `gorm:"column:artist_id" json:"artist_id"`
	Name      string `gorm:"column:name" json:"name"`
	SortOrder int    `gorm:"column:sort_order" json:"sort_order"`
}

func (*MusicArtistAlbum) TableName() string {
	return "music_artists_albums"
}
