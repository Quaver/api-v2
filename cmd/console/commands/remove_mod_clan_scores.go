package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RemoveUnrankedClanScores = &cobra.Command{
	Use:   "clan:remove:unranked",
	Short: "Removes unranked clan scores",
	Run: func(cmd *cobra.Command, args []string) {
		scores := make([]*db.Score, 0)

		result := db.SQL.
			Where("clan_id IS NOT NULL AND mods > 0").
			Find(&scores)

		if result.Error != nil {
			logrus.Error(result.Error)
			return
		}

		for _, score := range scores {
			if enums.IsModComboRanked(enums.Mods(score.Modifiers)) {
				continue
			}

			err := db.SQL.Model(&db.Score{}).
				Where("id = ?", score.Id).
				Update("clan_id", nil).Error

			if err != nil {
				logrus.Error(err)
				return
			}

			logrus.Info("Removed clan score: ", score.Id)
		}
	},
}
