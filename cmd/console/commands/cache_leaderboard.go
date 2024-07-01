package commands

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"strconv"

	"github.com/spf13/cobra"
)

var CacheLeaderboardCmd = &cobra.Command{
	Use:   "cache:leaderboard",
	Short: "Populates the leaderboards in cache",
	Run: func(cmd *cobra.Command, args []string) {
		// Clear previous cached leaderboards
		err := deleteOldLeaderboards()
		if err != nil {
			return
		}

		logrus.Println("Populating leaderboards...")

		err = populateLeaderboard()
		if err != nil {
			return
		}

		logrus.Println("Leaderboards populated.")
	},
}

func populateLeaderboard() error {
	batchSize := 1000
	offset := 0

	for {
		var users = make([]*db.User, 0)

		result := db.SQL.
			Joins("StatsKeys4").
			Joins("StatsKeys7").
			Where("allowed = 1").
			Limit(batchSize).
			Offset(offset).
			Order("id ASC").
			Find(&users)

		if result.Error != nil {
			return result.Error
		}

		if len(users) == 0 {
			break
		}

		processUsers(users)

		offset += batchSize
	}

	return nil
}

func processUsers(users []*db.User) {
	for _, user := range users {
		db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:leaderboard:%v", enums.GameModeKeys4), redis.Z{
			Score:  user.StatsKeys4.OverallPerformanceRating,
			Member: strconv.Itoa(user.Id),
		})

		db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:leaderboard:%v", enums.GameModeKeys7), redis.Z{
			Score:  user.StatsKeys7.OverallPerformanceRating,
			Member: strconv.Itoa(user.Id),
		})

		if user.Country != "XX" {
			db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:country_leaderboard:%v:%v", user.Country, enums.GameModeKeys4), redis.Z{
				Score:  user.StatsKeys4.OverallPerformanceRating,
				Member: strconv.Itoa(user.Id),
			})

			db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:country_leaderboard:%v:%v", user.Country, enums.GameModeKeys7), redis.Z{
				Score:  user.StatsKeys7.OverallPerformanceRating,
				Member: strconv.Itoa(user.Id),
			})
		}

		// ToDo Total Hits Leaderboard
	}
}

func deleteOldLeaderboards() error {
	var cursor uint64

	keys, cursor, err := db.Redis.Scan(db.RedisCtx, cursor, fmt.Sprintf("quaver:leaderboard:*"), 0).Result()

	if err != nil && err != redis.Nil {
		logrus.Println(err)
	}

	if len(keys) > 0 {
		_, err := db.Redis.Del(db.RedisCtx, keys...).Result()

		if err != nil {
			logrus.Fatalf("Failed to DELETE keys: %v", err)
		}
	}

	return err
}
