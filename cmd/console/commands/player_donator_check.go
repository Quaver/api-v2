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
		batchSize := 1000
		offset := 0

		currentTime := time.Now()

		for {
			var users = make([]*db.User, 0)

			result := db.SQL.
				Where("allowed = 1 && donator_end_time != 0").
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

					if enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
						if err := user.UpdateUserGroups(user.UserGroups &^ enums.UserGroupDonator); err != nil {
							logrus.Println(err)
						}

						if err := user.UpdateDonatorEndTime(0); err != nil {
							logrus.Println(err)
						}

						// Add back donator if user is premium
						if userIsDiscordPremiumUser(user) {
							if err := user.UpdateUserGroups(user.UserGroups | enums.UserGroupDonator); err != nil {
								logrus.Println(err)
							}

							if err := user.UpdateDonatorEndTime(3600000); err != nil {
								logrus.Println(err)
							}
						}
					}
				}
			}

			offset += batchSize
		}
	},
}

func userIsDiscordPremiumUser(user *db.User) bool {
	resp, err := resty.New().R().Get(fmt.Sprintf("%v/donator/discord/check/%v", config.Instance.Discord.BotAPI, user.DiscordId))

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
