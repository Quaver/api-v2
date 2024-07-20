package commands

import (
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

var PlayerDonatorCheckCmd = &cobra.Command{
	Use:   "player:donator:check",
	Short: "Checks if a player is a donator",
	Run: func(cmd *cobra.Command, args []string) {
		expiredUsers := map[int]*db.UserNotification{}

		removeExpiredDonator(expiredUsers)
		addDiscordPremiumDonator(expiredUsers)

		for _, notification := range expiredUsers {
			if err := notification.Insert(); err != nil {
				logrus.Error("Error inserting donator expired notification: ", err)
			}
		}
	},
}

func removeExpiredDonator(expiredUsers map[int]*db.UserNotification) {
	batchSize := 1000
	offset := 0

	currentTime := time.Now()

	for {
		var users = make([]*db.User, 0)

		result := db.SQL.
			Where("? & usergroups != 0", enums.UserGroupDonator).
			Limit(batchSize).
			Offset(offset).
			Find(&users)

		if result.Error != nil {
			logrus.Println(result.Error)
		}

		if len(users) == 0 {
			break
		}

		for _, user := range users {
			if user.DonatorEndTime < currentTime.UnixMilli() {
				logrus.Printf("User %s donator expired", user.Username)

				user.UserGroups = user.UserGroups &^ enums.UserGroupDonator

				result := db.SQL.Model(&db.User{}).
					Where("id = ?", user.Id).
					Update("usergroups", user.UserGroups).
					Update("donator_end_time", 0)

				if result.Error != nil {
					logrus.Println(result.Error)
				}

				expiredUsers[user.Id] = db.NewDonatorExpiredNotification(user.Id)
			}
		}

		offset += batchSize
	}
}

func addDiscordPremiumDonator(expiredUsers map[int]*db.UserNotification) {
	batchSize := 1000
	offset := 0

	for {
		var users = make([]*db.User, 0)

		result := db.SQL.
			Where("(? & usergroups = 0) AND discord_id IS NOT NULL", enums.UserGroupDonator).
			Limit(batchSize).
			Offset(offset).
			Find(&users)

		if result.Error != nil {
			logrus.Println(result.Error)
		}

		if len(users) == 0 {
			break
		}

		for _, user := range users {
			if userIsDiscordPremiumUser(user) {
				user.UserGroups = user.UserGroups | enums.UserGroupDonator

				result := db.SQL.Model(&db.User{}).
					Where("id = ?", user.Id).
					Update("usergroups", user.UserGroups).
					Update("donator_end_time", time.Now().UnixMilli()+3600000)

				if result.Error != nil {
					logrus.Println(result.Error)
				}

				logrus.Info("Gave donator to Discord premium user: ", user.Username)
				delete(expiredUsers, user.Id)
			}
		}
		offset += batchSize
	}
}

func userIsDiscordPremiumUser(user *db.User) bool {
	if user.DiscordId == nil {
		return false
	}

	resp, err := resty.New().R().Get(fmt.Sprintf("%v/donator/discord/check/%v", config.Instance.Discord.BotAPI, *user.DiscordId))

	if err != nil {
		return false
	}

	if resp.IsError() {
		return false
	}

	type response struct {
		HasDonator bool `json:"has_donator"`
	}

	var parsed response

	if err = json.Unmarshal(resp.Body(), &parsed); err != nil {
		return false
	}

	return parsed.HasDonator
}
