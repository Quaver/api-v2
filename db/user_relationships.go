package db

type UserRelationship struct {
	Id           int  `gorm:"column:id"`
	UserId       int  `gorm:"column:user_id"`
	TargetUserId int  `gorm:"column:target_user_id"`
	Relationship int8 `gorm:"column:relationship"`
}

func (*UserRelationship) TableName() string {
	return "user_relationships"
}

// GetUserRelationship Retrieves a user relationship in the database
func GetUserRelationship(userId int, targetUserId int) (*UserRelationship, error) {
	var relationship *UserRelationship

	result := SQL.
		Where("user_id = ? AND target_user_id = ?", userId, targetUserId).
		First(&relationship)

	if result.Error != nil {
		return nil, result.Error
	}

	return relationship, nil
}

// AddFriend Adds a friend to the database
func AddFriend(userId int, targetUserId int) error {
	relationship := UserRelationship{
		UserId:       userId,
		TargetUserId: targetUserId,
		Relationship: 1,
	}

	if err := SQL.Create(&relationship).Error; err != nil {
		return err
	}

	return nil
}

// RemoveFriend Removes a friend from the database
func RemoveFriend(userId int, targetUserId int) error {
	if err := SQL.Delete(&UserRelationship{},
		"user_id = ? AND target_user_id = ?", userId, targetUserId).Error; err != nil {
		return err
	}

	return nil
}
