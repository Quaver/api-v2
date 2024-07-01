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
		err := deleteOldLeaderboards()

		if err != nil {
			logrus.Error(err)
			return
		}

		logrus.Println("Populating leaderboards...")

		err = populateLeaderboard()

		if err != nil {
			logrus.Error(err)
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

		if err := processUsers(users); err != nil {
			return err
		}

		offset += batchSize
	}

	return nil
}

func processUsers(users []*db.User) error {
	for _, user := range users {
		err := db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:leaderboard:%v", enums.GameModeKeys4),
			redis.Z{
				Score:  user.StatsKeys4.OverallPerformanceRating,
				Member: strconv.Itoa(user.Id),
			}).Err()

		if err != nil {
			return err
		}

		err = db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:leaderboard:%v", enums.GameModeKeys7),
			redis.Z{
				Score:  user.StatsKeys7.OverallPerformanceRating,
				Member: strconv.Itoa(user.Id),
			}).Err()

		if user.Country != "XX" {
			err = db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:country_leaderboard:%v:%v",
				user.Country, enums.GameModeKeys4), redis.Z{
				Score:  user.StatsKeys4.OverallPerformanceRating,
				Member: strconv.Itoa(user.Id),
			}).Err()

			if err != nil {
				return err
			}

			err = db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:country_leaderboard:%v:%v",
				user.Country, enums.GameModeKeys7), redis.Z{
				Score:  user.StatsKeys7.OverallPerformanceRating,
				Member: strconv.Itoa(user.Id),
			}).Err()

			if err != nil {
				return err
			}
		}

		var userKeys4Hits = user.StatsKeys4.TotalMarvelous + user.StatsKeys4.TotalPerfect +
			user.StatsKeys4.TotalGreat + user.StatsKeys4.TotalGood + user.StatsKeys4.TotalOkay
		var userKeys7Hits = user.StatsKeys7.TotalMarvelous + user.StatsKeys7.TotalPerfect +
			user.StatsKeys7.TotalGreat + user.StatsKeys7.TotalGood + user.StatsKeys7.TotalOkay

		err = db.Redis.ZAdd(db.RedisCtx, fmt.Sprintf("quaver:leaderboard:total_hits_global"),
			redis.Z{
				Score:  float64(userKeys4Hits) + float64(userKeys7Hits),
				Member: strconv.Itoa(user.Id),
			}).Err()

		if err != nil {
			return err
		}
	}

	return nil
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
