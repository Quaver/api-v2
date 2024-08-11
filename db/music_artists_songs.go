package db

type MusicArtistSong struct {
	Id        int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	AlbumId   int    `gorm:"column:album_id" json:"album_id"`
	Name      string `gorm:"column:name" json:"name"`
	SortOrder int    `gorm:"column:sort_order" json:"sort_order"`
	Length    int    `gorm:"column:length" json:"length"`
	BPM       int    `gorm:"column:bpm" json:"bpm"`
}

func (*MusicArtistSong) TableName() string {
	return "music_artists_songs"
}
