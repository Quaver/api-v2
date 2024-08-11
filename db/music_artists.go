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
		Where("id = ?", id).
		First(&artist)

	if result.Error != nil {
		return nil, result.Error
	}

	return &artist, nil
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
