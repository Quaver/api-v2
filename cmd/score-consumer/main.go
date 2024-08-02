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

			db.Redis.XAck(db.RedisCtx, subject, consumersGroup, messageID)
		}
	}
}
