package db

type MusicArtist struct {
	Id            int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	Name          string `gorm:"column:name" json:"name"`
	Description   string `gorm:"column:description" json:"description"`
	ExternalLinks string `gorm:"column:external_links" json:"-"`
	SortOrder     int    `gorm:"column:sort_order" json:"sort_order"`
	Visible       bool   `gorm:"column:visible" json:"-"`
}

func (*MusicArtist) TableName() string {
	return "music_artists"
}

func (ma *MusicArtist) Insert() error {
	return SQL.Create(ma).Error
}
