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

func (album *MusicArtistAlbum) ID() int {
	return album.Id
}

// GetMusicArtistAlbums Retrieves a given music artist's albums
func GetMusicArtistAlbums(artistId int) ([]*MusicArtistAlbum, error) {
	albums := make([]*MusicArtistAlbum, 0)

	result := SQL.
		Where("artist_id = ?", artistId).
		Find(&albums)

	if result.Error != nil {
		return nil, result.Error
	}

	return albums, nil
}

// GetMusicArtistAlbumById Retrieves a music artist album from the db
func GetMusicArtistAlbumById(id int) (*MusicArtistAlbum, error) {
	var album MusicArtistAlbum

	result := SQL.
		Where("id = ?", id).
		First(&album)

	if result.Error != nil {
		return nil, result.Error
	}

	return &album, nil
}

// UpdateName Updates the name of the album
func (album *MusicArtistAlbum) UpdateName(name string) error {
	album.Name = name

	return SQL.
		Model(&MusicArtistAlbum{}).
		Where("id = ?", album.Id).
		Update("name", album.Name).
		Error
}

// Delete Deletes an album
func (album *MusicArtistAlbum) Delete() error {
	return SQL.
		Delete(&MusicArtistAlbum{}, "id = ?", album.Id).
		Error
}

// UpdateSortOrder Updates the sort order of an an album
func (album *MusicArtistAlbum) UpdateSortOrder(sortOrder int) error {
	album.SortOrder = sortOrder

	return SQL.Model(&MusicArtistAlbum{}).
		Where("id = ?", album.Id).
		Update("sort_order", album.SortOrder).Error
}

// SyncMusicArtistAlbumSortOrders Syncs the sort order of music artist albums
func SyncMusicArtistAlbumSortOrders(artistId int) error {
	artists, err := GetMusicArtistAlbums(artistId)

	if err != nil {
		return err
	}

	return SyncSortOrder(artists, func(album *MusicArtistAlbum, sortOrder int) error {
		return album.UpdateSortOrder(sortOrder)
	})
}
