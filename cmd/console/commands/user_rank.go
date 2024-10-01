package commands

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

var UserRankCmd = &cobra.Command{
	Use:   "stats:rank",
	Short: "Inserts the rank stats for all users ",
	Run: func(cmd *cobra.Command, args []string) {
		batchSize := 1000
		offset := 0

		for {
			var users = make([]*db.User, 0)

			result := db.SQL.
				Where("allowed = 1").
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
				for i := 1; i <= 2; i++ {
					key := fmt.Sprintf("quaver:leaderboard:%v", i)
					userStr := fmt.Sprintf("%v", user.Id)

					data, err := db.Redis.ZRevRankWithScore(db.RedisCtx, key, userStr).Result()

					if err != nil && err != redis.Nil {
						logrus.Error(err)
						return
					}

					if err == redis.Nil {
						logrus.Info("Skipping user: ", user.Id, " (no rank found)")
						continue
					}

					switch enums.GameMode(i) {
					case enums.GameModeKeys4:
						rank := &db.UserRankKeys4{
							UserId:                   user.Id,
							Rank:                     int(data.Rank + 1),
							OverallPerformanceRating: data.Score,
							Timestamp:                time.Now(),
						}

						if err := db.SQL.Create(&rank).Error; err != nil {
							logrus.Error(err)
							return
						}

					case enums.GameModeKeys7:
						rank := &db.UserRankKeys7{
							UserId:                   user.Id,
							Rank:                     int(data.Rank + 1),
							OverallPerformanceRating: data.Score,
							Timestamp:                time.Now(),
						}

						if err := db.SQL.Create(&rank).Error; err != nil {
							logrus.Error(err)
							return
						}
					}

					logrus.Info("Inserted rank for user: ", user.Id)
				}
			}

			offset += batchSize
		}

		for i := 1; i <= 2; i++ {
			table := fmt.Sprintf("user_rank_%v", enums.GetGameModeString(enums.GameMode(i)))
			query := fmt.Sprintf("DELETE FROM %v WHERE timestamp <= (CURDATE() - INTERVAL 90 DAY)", table)

			result := db.SQL.Exec(query)

			if result.Error != nil {
				logrus.Error(result.Error)
				return
			}
		}

		logrus.Info("Complete!")
	},
}
