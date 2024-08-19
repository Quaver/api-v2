package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var DenyOnHoldCmd = &cobra.Command{
	Use:   "ranking:queue:hold:deny",
	Short: "Denies mapsets in the ranking queue that are on hold for a month",
	Run: func(cmd *cobra.Command, args []string) {
		onHoldMapsets, err := db.GetOldOnHoldMapsetsInRankingQueue()

		if err != nil {
			logrus.Error(err)
			return
		}

		for _, mapset := range onHoldMapsets {
			if err := mapset.UpdateStatus(db.RankingQueueDenied); err != nil {
				logrus.Error(err)
				return
			}

			logrus.Info("Auto-denied on hold mapset: ", mapset.MapsetId)

			avatar := webhooks.QuaverLogo

			_ = webhooks.SendQueueWebhook(&db.User{
				Id:        2,
				Username:  "QuaverBot",
				AvatarUrl: &avatar,
			}, mapset.Mapset, db.RankingQueueActionDeny)
		}
	},
}
