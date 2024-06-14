package db

import (
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Id                          int               `gorm:"column:id; PRIMARY_KEY" json:"id"`
	SteamId                     string            `gorm:"column:steam_id" json:"steam_id"`
	Username                    string            `gorm:"column:username" json:"username"`
	TimeRegistered              int64             `gorm:"column:time_registered" json:"-"`
	TimeRegisteredJSON          time.Time         `gorm:"-:all" json:"time_registered"`
	Allowed                     bool              `gorm:"column:allowed" json:"allowed"`
	Privileges                  enums.Privileges  `gorm:"column:privileges" json:"privileges"`
	UserGroups                  enums.UserGroups  `gorm:"column:usergroups" json:"usergroups"`
	MuteEndTime                 int64             `gorm:"column:mute_endtime" json:"-"`
	MuteEndTimeJSON             time.Time         `gorm:"-:all" json:"mute_end_time"`
	LatestActivity              int64             `gorm:"column:latest_activity" json:"-"`
	LatestActivityJSON          time.Time         `gorm:"-:all" json:"latest_activity"`
	Country                     string            `gorm:"column:country" json:"country"`
	IP                          string            `gorm:"column:ip" json:"-"`
	AvatarUrl                   *string           `gorm:"column:avatar_url" json:"avatar_url"`
	Twitter                     *string           `gorm:"column:twitter" json:"twitter"`
	Title                       *string           `gorm:"column:title" json:"title"`
	CheckedPreviousAchievements bool              `gorm:"column:checked_previous_achievements" json:"-"`
	UserPage                    *string           `gorm:"column:userpage" json:"userpage"`
	TwitchUsername              *string           `gorm:"column:twitch_username" json:"twitch_username"`
	DonatorEndTime              int64             `gorm:"column:donator_end_time" json:"-"`
	DonatorEndTimeJSON          time.Time         `gorm:"-:all" json:"donator_end_time"`
	Notes                       *string           `gorm:"column:notes" json:"-"`
	DiscordId                   *string           `gorm:"column:discord_id" json:"discord_id"`
	Information                 *string           `gorm:"column:information" json:"-"`
	MiscInformation             *UserInformation  `gorm:"-:all" json:"misc_information"`
	UserPageDisabled            bool              `gorm:"column:userpage_disabled" json:"-"`
	ClanId                      *int              `gorm:"column:clan_id" json:"clan_id"`
	ClanLeaveTime               int64             `gorm:"column:clan_leave_time" json:"-"`
	ClanLeaveTimeJSON           time.Time         `gorm:"-:all" json:"clan_leave_time"`
	ShadowBanned                bool              `gorm:"column:shadow_banned" json:"-"`
	ClientStatus                *UserClientStatus `gorm:"-:all" json:"client_status,omitempty"`
	StatsKeys4                  *UserStatsKeys4   `gorm:"foreignKey:UserId" json:"stats_keys4,omitempty"`
	StatsKeys7                  *UserStatsKeys7   `gorm:"foreignKey:UserId" json:"stats_keys7,omitempty"`
}

type UserClientStatus struct {
	Status  int    `json:"status"`
	Mode    int    `json:"mode"`
	Content string `json:"content"`
}

type UserInformation struct {
	Discord             string         `json:"discord,omitempty"`
	Twitter             string         `json:"twitter,omitempty"`
	Twitch              string         `json:"twitch,omitempty"`
	Youtube             string         `json:"youtube,omitempty"`
	NotifyMapsetActions bool           `json:"notif_action_mapset,omitempty"`
	DefaultMode         enums.GameMode `json:"default_mode,omitempty"`
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

	if status, err := GetUserClientStatus(u.Id); err == nil {
		u.ClientStatus = status
	}

	if keys4Ranks, err := GetUserRanksForMode(u, enums.GameModeKeys4); err == nil && u.StatsKeys4 != nil {
		u.StatsKeys4.Ranks = keys4Ranks
	}

	if keys7Ranks, err := GetUserRanksForMode(u, enums.GameModeKeys7); err == nil && u.StatsKeys7 != nil {
		u.StatsKeys7.Ranks = keys7Ranks
	}

	if u.Information != nil {
		if err := json.Unmarshal([]byte(*u.Information), &u.MiscInformation); err != nil {
			return err
		}
	}

	return nil
}

// Insert Inserts a new user to the database
func (u *User) Insert() error {
	err := SQL.Transaction(func(tx *gorm.DB) error {
		// Insert User
		result := tx.Create(&u)

		if result.Error != nil {
			return result.Error
		}

		// Insert 4K User Stats
		if err := tx.Create(&UserStatsKeys4{UserId: u.Id}).Error; err != nil {
			return err
		}

		// Insert 7K User Stats
		if err := tx.Create(&UserStatsKeys7{UserId: u.Id}).Error; err != nil {
			return err
		}

		// Insert Activity Feed
		if err := tx.Create(&UserActivity{
			UserId:    u.Id,
			Type:      UserActivityRegistered,
			Timestamp: time.Now().UnixMilli(),
			MapsetId:  -1,
		}).Error; err != nil {
			return err
		}

		// Global / CountryCode Leaderboards
		for i := 1; i < 2; i++ {
			if err := Redis.ZAdd(RedisCtx, fmt.Sprintf("quaver:leaderboard:%v", i), redis.Z{
				Score:  0,
				Member: strconv.Itoa(u.Id),
			}).Err(); err != nil {
				return err
			}

			countryLb := fmt.Sprintf("quaver:country_leaderboard:%v:%v", strings.ToLower(u.Country), i)

			if err := Redis.ZAdd(RedisCtx, countryLb, redis.Z{
				Score:  0,
				Member: strconv.Itoa(u.Id),
			}).Err(); err != nil {
				return err
			}
		}

		// Total Hits Leaderboard
		if err := Redis.ZAdd(RedisCtx, "quaver:leaderboard:total_hits_global", redis.Z{
			Score:  0,
			Member: strconv.Itoa(u.Id),
		}).Err(); err != nil {
			return err
		}

		// Increment total user count
		if err := Redis.Incr(RedisCtx, "quaver:total_user").Err(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// CanJoinClan Returns if the user is eligible to join a new clan
func (u *User) CanJoinClan() bool {
	return u.ClanId == nil
}

// IsTrialRankingSupervisor Returns if the user is a trial ranking supervisor
func (u *User) IsTrialRankingSupervisor() bool {
	return strings.Contains(*u.Title, "Trial 4K") || strings.Contains(*u.Title, "Trial 7K")
}

// GetUserById Retrieves a user from the database by their user id
func GetUserById(id int) (*User, error) {
	var user *User

	result := SQL.
		Joins("StatsKeys4").
		Joins("StatsKeys7").
		Where("id = ?", id).
		First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// GetUserBySteamId Retrieves a user from the database by their steam id
func GetUserBySteamId(id string) (*User, error) {
	var user *User

	result := SQL.
		Joins("StatsKeys4").
		Joins("StatsKeys7").
		Where("steam_id = ?", id).
		First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// GetUserByUsername Retrieves a user from the database by their username
func GetUserByUsername(username string) (*User, error) {
	var user *User

	result := SQL.
		Joins("StatsKeys4").
		Joins("StatsKeys7").
		Where("username = ?", username).
		First(&user)

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
		Where("users.clan_id = ?", clanId).
		Find(&users)

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

// UpdateUserAboutMe Updates a user's about me
func UpdateUserAboutMe(userId int, aboutMe string) error {
	result := SQL.Model(&User{}).Where("id = ?", userId).Update("userpage", aboutMe)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// UpdateUserUsername Updates a user's username
func UpdateUserUsername(userId int, username string) error {
	result := SQL.Model(&User{}).Where("id = ?", userId).Update("username", username)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// UpdateUserAllowed Updates whether the user is allowed to play (banned)
func UpdateUserAllowed(userId int, isAllowed bool) error {
	result := SQL.Model(&User{}).Where("id = ?", userId).Update("allowed", isAllowed)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetUserClientStatus Retrieves a user's client status from Redis
func GetUserClientStatus(id int) (*UserClientStatus, error) {
	result, err := Redis.Get(RedisCtx, fmt.Sprintf("quaver:server:user_status:%v", id)).Result()

	if err != nil && err != redis.Nil {
		logrus.Error("Error getting user status from redis", err)
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	type redisClientStatus struct {
		Status  string `json:"s"`
		Mode    string `json:"m"`
		Content string `json:"c"`
	}

	var status *redisClientStatus

	if err := json.Unmarshal([]byte(result), &status); err != nil {
		logrus.Error("Error unmarshalling client status json", err)
		return nil, err
	}

	return &UserClientStatus{
		Status:  parseRedisIntWithDefault(status.Status, 0),
		Mode:    parseRedisIntWithDefault(status.Mode, 1),
		Content: status.Content,
	}, nil
}

// GetUserRanksForMode Retrieves a user's global and country ranks for a game mode
func GetUserRanksForMode(user *User, mode enums.GameMode) (*UserRanks, error) {
	global, err := getUserRank(user, fmt.Sprintf("quaver:leaderboard:%v", mode))

	if err != nil {
		logrus.Error("Error getting user global rank: ", err)
		return nil, err
	}

	country, err := getUserRank(user, fmt.Sprintf("quaver:country_leaderboard:%v:%v", strings.ToLower(user.Country), mode))

	if err != nil {
		logrus.Error("Error getting user country rank: ", err)
		return nil, err
	}

	totalHits, err := getUserRank(user, "quaver:leaderboard:total_hits_global")

	if err != nil {
		logrus.Error("Error getting user total hits rank: ", err)
		return nil, err
	}

	return &UserRanks{
		Global:    global,
		Country:   country,
		TotalHits: totalHits,
	}, nil
}

func getUserRank(user *User, key string) (int, error) {
	rank, err := Redis.ZRevRank(RedisCtx, key, strconv.Itoa(user.Id)).Result()

	if err != nil {
		if err == redis.Nil {
			return -1, nil
		}

		return -1, err
	}

	return int(rank) + 1, nil
}
