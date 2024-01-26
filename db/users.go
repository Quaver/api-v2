package db

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id                          int             `gorm:"column:id; PRIMARY_KEY" json:"id"`
	SteamId                     string          `gorm:"column:steam_id" json:"steam_id"`
	Username                    string          `gorm:"column:username" json:"username"`
	TimeRegistered              int64           `gorm:"column:time_registered" json:"-"`
	TimeRegisteredJSON          time.Time       `gorm:"-:all" json:"time_registered"`
	Allowed                     bool            `gorm:"column:allowed" json:"allowed"`
	Privileges                  int64           `gorm:"column:privileges" json:"privileges"`
	UserGroups                  int64           `gorm:"column:usergroups" json:"usergroups"`
	MuteEndTime                 int64           `gorm:"column:mute_endtime" json:"-"`
	MuteEndTimeJSON             time.Time       `gorm:"-:all" json:"mute_end_time"`
	LatestActivity              int64           `gorm:"column:latest_activity" json:"-"`
	LatestActivityJSON          time.Time       `gorm:"-:all" json:"latest_activity"`
	Country                     string          `gorm:"column:country" json:"country"`
	IP                          string          `gorm:"column:ip" json:"-"`
	AvatarUrl                   *string         `gorm:"column:avatar_url" json:"avatar_url"`
	Twitter                     *string         `gorm:"column:twitter" json:"twitter"`
	Title                       *string         `gorm:"column:title" json:"title"`
	CheckedPreviousAchievements bool            `gorm:"column:checked_previous_achievements" json:"-"`
	UserPage                    *string         `gorm:"column:userpage" json:"userpage"`
	TwitchUsername              *string         `gorm:"column:twitch_username" json:"twitch_username"`
	DonatorEndTime              int64           `gorm:"column:donator_end_time" json:"-"`
	DonatorEndTimeJSON          time.Time       `gorm:"-:all" json:"donator_end_time"`
	Notes                       *string         `gorm:"column:notes" json:"-"`
	DiscordId                   *string         `gorm:"column:discord_id" json:"discord_id"`
	Information                 *string         `gorm:"column:information" json:"social_media"`
	UserPageDisabled            bool            `gorm:"column:userpage_disabled" json:"-"`
	ClanId                      *int            `gorm:"column:clan_id" json:"clan_id"`
	ClanLeaveTime               int64           `gorm:"column:clan_leave_time" json:"-"`
	ClanLeaveTimeJSON           time.Time       `gorm:"-:all" json:"clan_leave_time"`
	ShadowBanned                bool            `gorm:"column:shadow_banned" json:"-"`
	StatsKeys4                  *UserStatsKeys4 `gorm:"foreignKey:UserId" json:"stats_keys4"`
	StatsKeys7                  *UserStatsKeys7 `gorm:"foreignKey:UserId" json:"stats_keys7"`
}

func (u *User) BeforeCreate(*gorm.DB) (err error) {
	t := time.Now()
	u.TimeRegisteredJSON = t
	u.MuteEndTimeJSON = t
	u.LatestActivityJSON = t
	u.DonatorEndTimeJSON = t
	u.ClanLeaveTimeJSON = t

	return nil
}

func (u *User) AfterFind(*gorm.DB) (err error) {
	u.TimeRegisteredJSON = time.UnixMilli(u.TimeRegistered)
	u.MuteEndTimeJSON = time.UnixMilli(u.MuteEndTime)
	u.LatestActivityJSON = time.UnixMilli(u.LatestActivity)
	u.DonatorEndTimeJSON = time.UnixMilli(u.DonatorEndTime)
	u.ClanLeaveTimeJSON = time.UnixMilli(u.ClanLeaveTime)

	return nil
}

// CanJoinClan Returns if the user is eligible to join a new clan
func (u *User) CanJoinClan() bool {
	return u.ClanId == nil
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

// GetUsersInClan Retrieves all the users that are in a given clan
func GetUsersInClan(clanId int) ([]*User, error) {
	var users []*User

	result := SQL.
		Joins("StatsKeys4").
		Joins("StatsKeys7").
		Where("users.clan_id = ?", clanId).Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// SearchUsersByName Searches for users that have a similar name to the query
func SearchUsersByName(searchQuery string) ([]*User, error) {
	var users []*User

	result := SQL.
		Joins("StatsKeys4").
		Joins("StatsKeys7").
		Where("username LIKE ? AND allowed = 1", fmt.Sprintf("%v%%", searchQuery)).
		Limit(50).
		Order("id ASC").
		Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// UpdateUserClan Updates a given user's clan in the database.
// Not passing in any clan id will set it to NULL.
func UpdateUserClan(userId int, clanId ...int) error {
	var clanIdVal *int = nil

	if len(clanId) > 0 {
		clanIdVal = &clanId[0]
	}

	result := SQL.Model(&User{}).Where("id = ?", userId).Update("clan_id", clanIdVal)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
