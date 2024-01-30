package db

type Badge struct {
	Id          int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`
}

func (b *Badge) TableName() string {
	return "badges"
}

type UserBadge struct {
	UserId  int `gorm:"column:user_id" json:"user_id"`
	BadgeId int `gorm:"column:badge_id" json:"badge_id"`
}

func (ub *UserBadge) TableName() string {
	return "user_badges"
}

// GetUserBadges Retrieves a user's badges from the database
func GetUserBadges(id int) ([]*Badge, error) {
	var badges []*Badge

	result := SQL.
		Joins("JOIN user_badges ON badges.id = user_badges.badge_id").
		Where("user_badges.user_id = ?", id).
		Find(&badges)

	if result.Error != nil {
		return nil, result.Error
	}

	return badges, nil
}
