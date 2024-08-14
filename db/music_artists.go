package db

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MusicArtist struct {
	Id                int             `gorm:"column:id; PRIMARY_KEY" json:"id"`
	Name              string          `gorm:"column:name" json:"name"`
	Description       string          `gorm:"column:description" json:"description"`
	ExternalLinks     string          `gorm:"column:external_links" json:"-"`
	ExternalLinksJSON json.RawMessage `gorm:"-:all" json:"external_links"`
	SortOrder         int             `gorm:"column:sort_order" json:"sort_order"`
	Visible           bool            `gorm:"column:visible" json:"-"`
}

func (*MusicArtist) TableName() string {
	return "music_artists"
}

func (ma *MusicArtist) Insert() error {
	return SQL.Create(ma).Error
}

func (ma *MusicArtist) ID() int {
	return ma.Id
}

func (ma *MusicArtist) AfterFind(*gorm.DB) error {
	if ma.ExternalLinks != "" {
		if err := json.Unmarshal([]byte(ma.ExternalLinks), &ma.ExternalLinksJSON); err != nil {
			logrus.Warningf("Cannot unmarshal external links json for music artist: #%v", ma.Id)
		}
	}

	return nil
}

// GetMusicArtistById Retrieves a music artist by their id
func GetMusicArtistById(id int) (*MusicArtist, error) {
	var artist MusicArtist

	result := SQL.
		Where("id = ? AND visible = 1", id).
		First(&artist)

	if result.Error != nil {
		return nil, result.Error
	}

	return &artist, nil
}

// GetMusicArtists Retrieves all music artists
func GetMusicArtists() ([]*MusicArtist, error) {
	artists := make([]*MusicArtist, 0)

	result := SQL.
		Where("visible = 1").
		Order("sort_order ASC").
		Find(&artists)

	if result.Error != nil {
		return nil, result.Error
	}

	return artists, nil
}

// UpdateName Updates a music artist's name
func (ma *MusicArtist) UpdateName(name string) error {
	ma.Name = name

	return SQL.Model(&MusicArtist{}).
		Where("id = ?", ma.Id).
		Update("name", ma.Name).Error
}

// UpdateDescription Updates a music artist's description
func (ma *MusicArtist) UpdateDescription(description string) error {
	ma.Description = description

	return SQL.Model(&MusicArtist{}).
		Where("id = ?", ma.Id).
		Update("description", ma.Description).Error
}

// UpdateExternalLinks Updates the external links for a music artist. Must be valid JSON
func (ma *MusicArtist) UpdateExternalLinks(externalLinks string) error {
	if err := json.Unmarshal([]byte(externalLinks), &ma.ExternalLinksJSON); err != nil {
		return err
	}

	ma.ExternalLinks = externalLinks

	return SQL.Model(&MusicArtist{}).
		Where("id = ?", ma.Id).
		Update("external_links", ma.ExternalLinks).Error
}

// UpdateVisibility Updates the visibility of a music artist
func (ma *MusicArtist) UpdateVisibility(visible bool) error {
	ma.Visible = visible

	return SQL.Model(&MusicArtist{}).
		Where("id = ?", ma.Id).
		Update("visible", ma.Visible).Error
}

// UpdateSortOrder Updates the sort order of a single music artist
func (ma *MusicArtist) UpdateSortOrder(sortOrder int) error {
	ma.SortOrder = sortOrder

	return SQL.Model(&MusicArtist{}).
		Where("id = ?", ma.Id).
		Update("sort_order", ma.SortOrder).Error
}

// SyncMusicArtistSortOrders Syncs the sort order of music artists
func SyncMusicArtistSortOrders() error {
	artists, err := GetMusicArtists()

	if err != nil {
		return err
	}

	return SyncSortOrder(artists, func(artist *MusicArtist, sortOrder int) error {
		return artist.UpdateSortOrder(sortOrder)
	})
}
