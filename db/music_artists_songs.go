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

// GetMusicArtistSongById Retrieves a song from the db by its id
func GetMusicArtistSongById(id int) (*MusicArtistSong, error) {
	var song MusicArtistSong

	result := SQL.
		Where("id = ?", id).
		First(&song)

	if result.Error != nil {
		return nil, result.Error
	}

	return &song, nil
}

func (song *MusicArtistSong) UpdateName(name string) error {
	song.Name = name

	return SQL.Model(&MusicArtistSong{}).
		Where("id = ?", song.Id).
		Update("name", song.Name).Error
}

func (song *MusicArtistSong) UpdateBPM(bpm int) error {
	song.BPM = bpm

	return SQL.Model(&MusicArtistSong{}).
		Where("id = ?", song.Id).
		Update("bpm", song.BPM).Error
}

func (song *MusicArtistSong) UpdateLength(length int) error {
	song.Length = length

	return SQL.Model(&MusicArtistSong{}).
		Where("id = ?", song.Id).
		Update("length", song.Length).Error
}
