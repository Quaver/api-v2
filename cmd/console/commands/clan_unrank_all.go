package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ClanUnrankAllCommand = &cobra.Command{
	Use:   "clan:unrank:all",
	Short: "Unranks all clan maps",
	Run: func(cmd *cobra.Command, args []string) {
		maps := make([]*db.MapQua, 0)

		result := db.SQL.
			Where("clan_ranked = 1").
			Find(&maps)

		if result.Error != nil {
			logrus.Error(result.Error)
			return
		}

		for _, mapQua := range maps {
			if err := db.UpdateMapClanRanked(mapQua.Id, false); err != nil {
				logrus.Error("Error updating map clan ranked: ", err)
				return
			}

			result := db.SQL.Model(&db.Score{}).
				Where("map_md5 = ?", mapQua.MD5).
				Update("clan_id", nil)

			if result.Error != nil {
				logrus.Error("Error resetting clan_id for scores: ", result.Error)
				return
			}

			result = db.SQL.Delete(&db.ClanScore{}, "map_md5 = ?", mapQua.MD5)

			if result.Error != nil {
				logrus.Error("Error deleting clan scores: ", result.Error)
				return
			}

			logrus.Info("Successfully unranked clan map: ", mapQua)
		}
	},
}
