package commands

import (
	"encoding/json"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ClanRankMapCmd = &cobra.Command{
	Use:   "clan:rank:map",
	Short: "Ranks a map for clans",
	Run: func(cmd *cobra.Command, args []string) {
		for i := 1; i <= 2; i++ {
			var mapQua *db.MapQua

			result := db.SQL.Raw("SELECT * FROM maps "+
				"WHERE maps.clan_ranked = 0 AND maps.ranked_status = 2 AND maps.game_mode = ? "+
				"ORDER BY RAND() LIMIT 1;", i).
				Scan(&mapQua)

			if result.Error != nil {
				logrus.Error("Error retrieving random map for clan ranking: ", result.Error)
				return
			}

			if err := db.UpdateMapClanRanked(mapQua.Id, true); err != nil {
				logrus.Error("Error updating clan ranked status: ", err)
				return
			}

			clanUsers, err := db.GetAllUsersInAClan()

			if err != nil {
				logrus.Error("Error retrieving users a part of a clan", err)
				return
			}

			for _, user := range clanUsers {
				if err := db.NewClanMapRankedNotification(mapQua, user.Id).Insert(); err != nil {
					logrus.Error("Error inserting clan map ranked notification", err)
					return
				}
			}

			if err := sendClanMapToRedis(mapQua); err != nil {
				logrus.Error("Error sending clan map to redis: ", err)
			}

			logrus.Info("Ranked Clan Map: ", mapQua.Id, mapQua)
			_ = webhooks.SendClanRankedWebhook(mapQua)
		}
	},
}

// Publishes a ranked clan map to redis
func sendClanMapToRedis(mapQua *db.MapQua) error {
	type payload struct {
		Map struct {
			Id             int    `json:"id"`
			Artist         string `json:"artist"`
			Title          string `json:"title"`
			DifficultyName string `json:"difficulty_name"`
			CreatorName    string `json:"creator_name"`
			Mode           string `json:"mode"`
		} `json:"map"`
	}

	data := payload{}
	data.Map.Id = mapQua.Id
	data.Map.Artist = mapQua.Artist
	data.Map.Title = mapQua.Title
	data.Map.DifficultyName = mapQua.DifficultyName
	data.Map.CreatorName = mapQua.CreatorUsername
	data.Map.Mode = enums.GetShorthandGameModeString(mapQua.GameMode)

	dataStr, _ := json.Marshal(data)

	if err := db.Redis.Publish(db.RedisCtx, "quaver:ranked_clan_map", dataStr).Err(); err != nil {
		return err
	}

	return nil
}
