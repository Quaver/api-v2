package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var FixStatsCmd = &cobra.Command{
	Use:   "stats:fix",
	Short: "Fixes missing stats on a user",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logrus.Error("You must provide a user id to fix")
		}

		id, err := strconv.Atoi(args[0])

		if err != nil {
			logrus.Error(err)
			return
		}

		user, err := db.GetUserById(id)

		if err != nil {
			logrus.Error(err)
			return
		}

		if user.StatsKeys4 == nil {
			if err := db.SQL.Create(&db.UserStatsKeys4{UserId: user.Id}).Error; err != nil {
				logrus.Error(err)
				return
			}

			logrus.Info("Fixed missing 4K stats for user: ", user.Id)
		}

		if user.StatsKeys7 == nil {
			if err := db.SQL.Create(&db.UserStatsKeys7{UserId: user.Id}).Error; err != nil {
				logrus.Error(err)
				return
			}

			logrus.Info("Fixed missing 7K stats for user: ", user.Id)
		}
	},
}