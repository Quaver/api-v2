package main

import (
	"encoding/json"
	"flag"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/webhooks"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "../../config.json", "path to config file")
	flag.Parse()

	if err := config.Load(*configPath); err != nil {
		logrus.Panic(err)
	}

	if !config.Instance.IsProduction {
		logrus.SetLevel(logrus.DebugLevel)
	}

	db.ConnectMySQL()
	db.InitializeRedis()
	db.InitializeElasticSearch()
	azure.InitializeClient()
	webhooks.InitializeWebhooks()

	go consumeScores()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Exiting...")
}

func consumeScores() {
	subject := "quaver:scores:stream"
	consumersGroup := "api-score-consumer-group"

	err := db.Redis.XGroupCreate(db.RedisCtx, subject, consumersGroup, "0").Err()

	if err != nil {
		logrus.Warn(err)
	}

	for {
		entries, err := db.Redis.XReadGroup(db.RedisCtx, &redis.XReadGroupArgs{
			Group:    consumersGroup,
			Consumer: "api-score-consumer",
			Streams:  []string{subject, ">"},
			Count:    2,
			Block:    0,
			NoAck:    false,
		}).Result()

		if err != nil {
			logrus.Fatal(err)
		}

		for i := 0; i < len(entries[0].Messages); i++ {
			messageID := entries[0].Messages[i].ID
			scoreStr := entries[0].Messages[i].Values["data"]

			var score db.RedisScore

			if err := json.Unmarshal([]byte(scoreStr.(string)), &score); err != nil {
				logrus.Error(err)
				break
			}

			logrus.Infof("New Score: %v (#%v) | Map #%v | Difficulty: %v | Score #%v | Rating: %v | Acc: %v%%",
				score.User.Username, score.User.Id, score.Map.Id, score.Map.DifficultyRating,
				score.Score.Id, score.Score.PerformanceRating, score.Score.Accuracy)

			if err := db.Redis.Incr(db.RedisCtx, "quaver:total_scores").Err(); err != nil {
				logrus.Error("Error incrementing total score count in Redis", err)
			}

			go func() {
				if err := insertClanScore(&score); err != nil {
					logrus.Error("Error inserting clan score: ", err)
				}
			}()

			db.Redis.XAck(db.RedisCtx, subject, consumersGroup, messageID)
			db.Redis.XDel(db.RedisCtx, subject, messageID)
		}
	}
}

// Handles the insertion of a new clan score
func insertClanScore(score *db.RedisScore) error {
	if !score.Map.ClanRanked || score.User.ClanId <= 0 {
		return nil
	}

	if score.Score.Failed {
		return nil
	}

	existingScore, err := db.GetClanScore(score.Map.MD5, score.User.ClanId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	scoreboard, err := db.GetClanScoreboardForMap(score.Map.MD5)

	if err != nil {
		return err
	}

	newScore, err := db.CalculateClanScore(score.Map.MD5, score.User.ClanId, score.Map.GameMode)

	if err != nil {
		return err
	}

	// Make sure the id is the same on the newly calculated score, so it can be upserted properly.
	if existingScore != nil {
		newScore.Id = existingScore.Id
	}

	if err := db.SQL.Save(&newScore).Error; err != nil {
		return err
	}

	if err := db.RecalculateClanStats(score.User.ClanId, score.Map.GameMode, score); err != nil {
		return err
	}

	// Set values needed for the webhook
	clan, err := db.GetClanById(score.User.ClanId)

	if err != nil {
		return err
	}

	mapQua, err := db.GetMapById(score.Map.Id)

	if err != nil {
		return err
	}

	if len(scoreboard) == 0 {
		_ = webhooks.SendClanFirstPlaceWebhook(clan, mapQua, newScore, nil)
	} else if scoreboard[0].ClanId != score.User.ClanId && newScore.OverallRating > scoreboard[0].OverallRating {
		_ = webhooks.SendClanFirstPlaceWebhook(clan, mapQua, newScore, scoreboard[0])
	}

	return nil
}
