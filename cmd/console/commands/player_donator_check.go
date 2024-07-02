package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
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
					}
				}
			}

			offset += batchSize
		}
	},
}
