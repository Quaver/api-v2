package commands

import (
	"github.com/Quaver/api2/db"
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
			}

			logrus.Info("Ranked Clan Map: ", mapQua.Id, mapQua)
			_ = webhooks.SendClanRankedWebhook(mapQua)
		}
	},
}
