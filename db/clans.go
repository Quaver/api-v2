package db

type Clan struct {
	Id                 int    `gorm:"column:id"`
	OwnerId            int    `gorm:"column:owner_id"`
	Name               string `gorm:"column:name"`
	Tag                string `gorm:"column:tag"`
	CreatedAt          int64  `gorm:"column:created_at"`
	AboutMe            string `gorm:"column:about_me"`
	FavoriteMode       uint8  `gorm:"column:favorite_mode"`
	LastNameChangeTime int64  `gorm:"column:last_name_change_time"`
}

// GetClanByName Gets a clan from the database by its name
func GetClanByName(name string) (*Clan, error) {
	var clan *Clan

	result := SQL.Where("name = ?", name).First(&clan)

	if result.Error != nil {
		return nil, result.Error
	}

	return clan, nil
}
