package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CacheClanLeaderboard = &cobra.Command{
	Use:   "cache:leaderboard:clan",
	Short: "Caches the clan leaderboard",
	Run: func(cmd *cobra.Command, args []string) {
		batchSize := 1000
		offset := 0

		logrus.Info("Caching clan leaderboards...")

		for {
			var clans = make([]*db.Clan, 0)

			result := db.SQL.
				Preload("Stats").
				Limit(batchSize).
				Offset(offset).
				Order("id ASC").
				Find(&clans)

			if result.Error != nil {
				logrus.Error(result.Error)
				return
			}

			if len(clans) == 0 {
				break
			}

			for _, clan := range clans {
				for i := 1; i <= 2; i++ {
					if err := db.UpdateClanLeaderboard(clan, enums.GameMode(i)); err != nil {
						logrus.Error(err)
						return
					}
				}
			}

			offset += batchSize
		}

		logrus.Info("Complete!")
	},
}
