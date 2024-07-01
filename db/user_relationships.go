package db

import "gorm.io/gorm"

type UserRelationship struct {
	Id           int   `gorm:"column:id" json:"-"`
	UserId       int   `gorm:"column:user_id" json:"-"`
	TargetUserId int   `gorm:"column:target_user_id" json:"-"`
	Relationship int8  `gorm:"column:relationship" json:"-"`
	User         *User `gorm:"foreignKey:TargetUserId" json:"-"`
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

type UserFriend struct {
	User
	IsMutual bool `json:"is_mutual"`
}

// GetUserFriends Returns a user's friends list
func GetUserFriends(userId int) ([]*UserFriend, error) {
	var relationships = make([]*UserRelationship, 0)

	result := SQL.
		Preload("User").
		Preload("User.StatsKeys4").
		Preload("User.StatsKeys7").
		Where("user_relationships.user_id = ? AND user_relationships.relationship = 1", userId).
		Find(&relationships)

	if result.Error != nil {
		return nil, result.Error
	}

	var friends = make([]*UserFriend, 0)

	for _, relationship := range relationships {
		friend := &UserFriend{
			User:     *relationship.User,
			IsMutual: false,
		}

		mutualRelationship, err := GetUserRelationship(relationship.User.Id, userId)

		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		friend.IsMutual = mutualRelationship != nil
		friends = append(friends, friend)
	}

	return friends, nil
}
