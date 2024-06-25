package db

import "gorm.io/gorm"

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

func (ub *UserBadge) Insert() error {
	if err := SQL.Create(&ub).Error; err != nil {
		return err
	}

	return nil
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

// UserHasBadge Returns if a user has a particular badge.
func UserHasBadge(userId int, badgeId int) (bool, error) {
	var badge *UserBadge

	result := SQL.
		Where("user_id = ? AND badge_id = ?", userId, badgeId).
		First(&badge)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}

		return false, result.Error
	}

	return true, nil
}
