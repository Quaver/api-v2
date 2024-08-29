package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var DatabaseScoresBatchDeleteFailed = &cobra.Command{
	Use:   "database:scores:batch:delete",
	Short: "Deletes every failed score",
	Run: func(cmd *cobra.Command, args []string) {
		batchSize := 5000
		offset := 0

		for {
			var scoreIDs []int

			result := db.SQL.
				Model(&db.Score{}).
				Where("failed = 1").
				Limit(batchSize).
				Offset(offset).
				Pluck("id", &scoreIDs)

			if result.Error != nil {
				logrus.Info("Something went wrong")
				break
			}

			if len(scoreIDs) == 0 {
				if offset == 0 {
					logrus.Info("No more failed scores found!")
				} else {
					logrus.Info("There are no scores left to delete!")
				}

				break
			}

			deletedBatch := db.SQL.Where("id IN ?", scoreIDs).Delete(&db.Score{})

			if deletedBatch.Error != nil {
				logrus.Info("Something went wrong when deleting scores")
				break
			}

			logrus.Infof("Deleted %d failed scores", deletedBatch.RowsAffected)

			offset += batchSize
		}
	},
}
