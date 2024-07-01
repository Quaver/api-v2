package db

import "gorm.io/gorm"

type Achievement struct {
	Id           int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	Difficulty   string `gorm:"column:difficulty" json:"difficulty"`
	SteamAPIName string `gorm:"column:steam_api_name" json:"steam_api_name"`
	Name         string `gorm:"column:name" json:"name"`
	Description  string `gorm:"column:description" json:"description"`
	IsUnlocked   bool   `gorm:"-:all" json:"is_unlocked"`
}

func (*Achievement) TableName() string {
	return "achievements"
}

type UserAchievement struct {
	UserId        int `gorm:"column:user_id" json:"user_id"`
	AchievementId int `gorm:"column_achievement_id" json:"achievement_id"`
}

func (*UserAchievement) TableName() string {
	return "user_achievements"
}

// GetUserAchievements Gets a user's unlocked achievements
func GetUserAchievements(id int) ([]*Achievement, error) {
	var achievements = make([]*Achievement, 0)
	result := SQL.Order("id ASC").Find(&achievements)

	if result.Error != nil {
		return nil, result.Error
	}

	var userAchievements = make([]*UserAchievement, 0)
	result = SQL.Where("user_id = ?", id).Find(&userAchievements)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	// Go through each achievement, and set its unlocked status
	for _, userAchievement := range userAchievements {
		for _, achievement := range achievements {
			if userAchievement.AchievementId == achievement.Id {
				achievement.IsUnlocked = true
			}
		}
	}

	return achievements, nil
}
