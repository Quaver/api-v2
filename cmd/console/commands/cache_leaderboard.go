package commands

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var CacheLeaderboardCmd = &cobra.Command{
	Use:   "cache:leaderboard",
	Short: "Populates the leaderboards in cache",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && strings.ToLower(args[0]) == "delete-all" {
			if err := deleteOldLeaderboards(); err != nil {
				logrus.Error(err)
				return
			}
		}

		logrus.Println("Populating leaderboards...")

		if err := populateLeaderboard(); err != nil {
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
		for i := 1; i < 2; i++ {
			mode := enums.GameMode(i)
			globalKey := fmt.Sprintf("quaver:leaderboard:%v", mode)
			countryKey := fmt.Sprintf("quaver:country_leaderboard:%v:%v", strings.ToLower(user.Country), mode)

			// User was banned so remove them from the global/country leaderboards
			if !user.Allowed {
				if err := db.Redis.ZRem(db.RedisCtx, globalKey, strconv.Itoa(user.Id)).Err(); err != nil {
					return err
				}

				if err := db.Redis.ZRem(db.RedisCtx, countryKey, strconv.Itoa(user.Id)).Err(); err != nil {
					return err
				}

				continue
			}

			err := db.Redis.ZAdd(db.RedisCtx, globalKey, redis.Z{
				Score:  user.StatsKeys4.OverallPerformanceRating,
				Member: strconv.Itoa(user.Id),
			}).Err()

			if err != nil {
				return err
			}

			if user.Country != "XX" {
				err = db.Redis.ZAdd(db.RedisCtx, countryKey, redis.Z{
					Score:  user.StatsKeys4.OverallPerformanceRating,
					Member: strconv.Itoa(user.Id),
				}).Err()

				if err != nil {
					return err
				}
			}
		}

		totalHitsKey := "quaver:leaderboard:total_hits_global"

		// User was banned, so remove them from the total hits leaderboard
		if !user.Allowed {
			if err := db.Redis.ZRem(db.RedisCtx, totalHitsKey, strconv.Itoa(user.Id)).Err(); err != nil {
				return err
			}

			continue
		}

		var userKeys4Hits = user.StatsKeys4.TotalMarvelous + user.StatsKeys4.TotalPerfect +
			user.StatsKeys4.TotalGreat + user.StatsKeys4.TotalGood + user.StatsKeys4.TotalOkay
		var userKeys7Hits = user.StatsKeys7.TotalMarvelous + user.StatsKeys7.TotalPerfect +
			user.StatsKeys7.TotalGreat + user.StatsKeys7.TotalGood + user.StatsKeys7.TotalOkay

		err := db.Redis.ZAdd(db.RedisCtx, totalHitsKey, redis.Z{
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
	logrus.Info("Deleting old leaderboards...")

	keys, err := db.Redis.Keys(db.RedisCtx, "quaver:leaderboard:*").Result()

	if err != nil && err != redis.Nil {
		logrus.Println(err)
	}

	countryKeys, err := db.Redis.Keys(db.RedisCtx, "quaver:country_leaderboard:*").Result()

	if err != nil && err != redis.Nil {
		logrus.Println(err)
	}

	keys = append(keys, countryKeys...)

	if len(keys) > 0 {
		_, err := db.Redis.Del(db.RedisCtx, keys...).Result()

		if err != nil {
			logrus.Errorf("Failed to DELETE keys: %v", err)
			return err
		}
	}

	return err
}
