package db

import (
	"time"
)

type User struct {
	Id                          int     `gorm:"column:id; PRIMARY_KEY"`
	SteamId                     string  `gorm:"column:steam_id"`
	Username                    string  `gorm:"column:username"`
	TimeRegistered              int64   `gorm:"column:time_registered"`
	Allowed                     bool    `gorm:"column:allowed"`
	Privileges                  int64   `gorm:"column:privileges"`
	UserGroups                  int64   `gorm:"column:usergroups"`
	MuteEndTime                 int64   `gorm:"column:mute_endtime"`
	LatestActivity              int64   `gorm:"column:latest_activity"`
	Country                     string  `gorm:"column:country"`
	IP                          string  `gorm:"column:ip"`
	AvatarUrl                   *string `gorm:"column:avatar_url"`
	Twitter                     *string `gorm:"column:twitter"`
	Title                       *string `gorm:"column:title"`
	CheckedPreviousAchievements bool    `gorm:"column:checked_previous_achievements"`
	UserPage                    *string `gorm:"column:userpage"`
	TwitchUsername              *string `gorm:"column:twitch_username"`
	DonatorEndTime              int64   `gorm:"column:donator_end_time"`
	Notes                       *string `gorm:"column:notes"`
	DiscordId                   *string `gorm:"column:discord_id"`
	Information                 *string `gorm:"column:information"`
	UserPageDisabled            bool    `gorm:"column:userpage_disabled"`
	ClanId                      *int    `gorm:"column:clan_id"`
	ClanLeaveTime               int64   `gorm:"column:clan_leave_time"`
	ShadowBanned                bool    `gorm:"column:shadow_banned"`
}

// CanJoinClan Returns if the user is eligible to join a new clan
func (u *User) CanJoinClan() bool {
	return u.ClanId == nil && time.Now().Sub(time.UnixMilli(u.ClanLeaveTime)) >= (time.Hour*24)
}

// GetUserById Retrieves a user from the database by their Steam Id
func GetUserById(id int) (*User, error) {
	var user *User
	result := SQL.Where("id = ?", id).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// UpdateUserClan Updates a given user's clan in the database.
func UpdateUserClan(userId int, clanId int) error { //	SQL.Update("clan_id = ?", clanId).Where()
	return nil
}
