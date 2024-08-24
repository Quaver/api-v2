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

func (song *MusicArtistSong) ID() int {
	return song.Id
}

// GetMusicArtistSongsInAlbum Returns the songs in a music artist's album
func GetMusicArtistSongsInAlbum(albumId int) ([]*MusicArtistSong, error) {
	songs := make([]*MusicArtistSong, 0)

	result := SQL.
		Where("album_id = ?", albumId).
		Order("sort_order ASC").
		Find(&songs)

	if result.Error != nil {
		return nil, result.Error
	}

	return songs, nil
}
