package main

import (
	"encoding/json"
	"flag"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
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

			if score.Score.Failed {
				if err := db.IncrementFailedScoresMetric(); err != nil {
					logrus.Error("Error incrementing failed score metric in db", err)
				}
			}

			if err := db.Redis.Incr(db.RedisCtx, "quaver:total_scores").Err(); err != nil {
				logrus.Error("Error incrementing total score count in Redis", err)
			}

			if err := insertClanScore(&score); err != nil {
				logrus.Error("Error inserting clan score: ", err)
			}

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

	if err := db.UpdateAllClanLeaderboards(clan); err != nil {
		return err
	}

	mapQua, err := db.GetMapById(score.Map.Id)

	if err != nil {
		return err
	}

	if err := handleClanFirstPlaces(score, clan, mapQua, newScore, scoreboard); err != nil {
		return err
	}

	return nil
}

func handleClanFirstPlaces(score *db.RedisScore, clan *db.Clan, mapQua *db.MapQua, newScore *db.ClanScore, scoreboard []*db.ClanScore) error {
	var achievedFirstPlace bool

	if len(scoreboard) == 0 {
		achievedFirstPlace = true
		_ = webhooks.SendClanFirstPlaceWebhook(clan, mapQua, newScore, nil)
	} else if scoreboard[0].ClanId != score.User.ClanId && newScore.OverallRating > scoreboard[0].OverallRating {
		achievedFirstPlace = true
		_ = webhooks.SendClanFirstPlaceWebhook(clan, mapQua, newScore, scoreboard[0])
	}

	if !achievedFirstPlace {
		return nil
	}

	// Add activity for clan who won first place
	firstPlaceActivity := db.NewClanActivity(clan.Id, db.ClanActivityAchievedFirstPlace, score.User.Id)
	firstPlaceActivity.MapId = mapQua.Id
	firstPlaceActivity.Message = mapQua.String()

	if err := firstPlaceActivity.Insert(); err != nil {
		return err
	}

	if err := SendClanFirstPLaceToRedis(clan.Id, true, mapQua); err != nil {
		return err
	}

	if len(scoreboard) == 0 {
		return nil
	}

	// Add activity for clan who lost first place
	lostActivity := db.NewClanActivity(scoreboard[0].ClanId, db.ClanActivityLostFirstPlace, score.User.Id)
	lostActivity.MapId = mapQua.Id
	lostActivity.Message = mapQua.String()

	if err := lostActivity.Insert(); err != nil {
		return err
	}

	if err := SendClanFirstPLaceToRedis(scoreboard[0].ClanId, false, mapQua); err != nil {
		return err
	}

	// Add lost notification for clan members
	clanMembers, err := db.GetUsersInClan(scoreboard[0].ClanId)

	if err != nil {
		return err
	}

	for _, member := range clanMembers {
		if err := db.NewClanLostFirstPlaceNotification(mapQua, member.Id).Insert(); err != nil {
			return err
		}
	}

	return nil
}

func SendClanFirstPLaceToRedis(clanId int, won bool, mapQua *db.MapQua) error {
	type payload struct {
		ClanId int  `json:"clan_id"`
		Won    bool `json:"won"`
		Map    struct {
			Id             int    `json:"id"`
			Artist         string `json:"artist"`
			Title          string `json:"title"`
			DifficultyName string `json:"difficulty_name"`
			CreatorName    string `json:"creator_name"`
			Mode           string `json:"mode"`
		} `json:"map"`
	}

	data := payload{}
	data.ClanId = clanId
	data.Won = won
	data.Map.Id = mapQua.Id
	data.Map.Artist = mapQua.Artist
	data.Map.Title = mapQua.Title
	data.Map.DifficultyName = mapQua.DifficultyName
	data.Map.CreatorName = mapQua.CreatorUsername
	data.Map.Mode = enums.GetShorthandGameModeString(mapQua.GameMode)

	dataStr, _ := json.Marshal(data)

	if err := db.Redis.Publish(db.RedisCtx, "quaver:clan_first_place", dataStr).Err(); err != nil {
		return err
	}

	return nil
}
